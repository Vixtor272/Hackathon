#!/usr/bin/env python3
"""End-to-end smoke test of the Farmi MVP.

Starts the Go backend on a scratch port and drives the full flow through the
REST contract (docs/API.md): WhatsApp → OCR → validation → availability →
cart → reservation → payment → fulfilment → notifications, for pickup,
delivery and the expiry/renew/cancel paths. Exit code 1 if any check fails.
"""
import json
import os
import pathlib
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[1]
PORT = int(os.environ.get("SMOKE_PORT", "18080"))
BASE = f"http://localhost:{PORT}/api/v1"
WEB = "http://localhost:5173"
A, B, C = "+593991111111", "+593992222222", "+593993333333"

checks = []


def check(name, cond):
    checks.append((name, bool(cond)))
    print(("  ✓ " if cond else "  ✗ ") + name)


def call(method, path, body=None):
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(BASE + path, data=data, method=method, headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=5) as r:
            raw = r.read()
            return r.status, (json.loads(raw) if raw else None)
    except urllib.error.HTTPError as e:
        raw = e.read()
        return e.code, (json.loads(raw) if raw else None)


def chat(phone, text=None, media=None):
    body = {"from": phone, "type": "image" if media else "text"}
    if media:
        body["mediaId"] = media
    else:
        body["text"] = text
    st, res = call("POST", "/whatsapp/webhook", body)
    assert st == 200, (st, res)
    return res["state"], "\n".join(m["text"] for m in res["replies"]), res["replies"]


def transcript(phone):
    return call("GET", f"/whatsapp/conversations/{urllib.parse.quote(phone)}/messages")[1]


def said(phone, needle):
    return any(needle in m["text"] for m in transcript(phone)["messages"])


def order_from(replies):
    link = next(m for m in replies if m["type"] == "link")
    return link["link"].rsplit("/", 1)[1]


def pay_card(order_id):
    st, res = call("POST", "/payments", {"orderId": order_id, "method": "card"})
    assert st == 201, res
    st, res = call("POST", f"/payments/{res['payment']['id']}/confirm", {"outcome": "approved"})
    assert st == 200, res
    return res


def pickup_flow():
    print("\n[1] Pickup flow, zona centro (split between two pharmacies)")
    s, r, _ = chat(A, "hola")
    check("greets and asks for the cédula first", s == "ASK_ID" and "cédula" in r)
    s, r, _ = chat(A, "1712345678")
    check("recognises María by cédula", s == "ASK_PRESCRIPTION" and "María" in r)
    s, r, _ = chat(A, media="receta-004")
    check("rejects prescription without signature/stamp", s == "ASK_PRESCRIPTION" and "firma" in r)
    s, r, _ = chat(A, media="receta-003")
    check("rejects inactive doctor", "no está activo" in r)
    s, r, _ = chat(A, media="receta-006")
    check("rejects unknown doctor", "no está registrado" in r)
    s, r, _ = chat(A, media="receta-005")
    check("reports illegible image", "ilegible" in r)
    s, r, _ = chat(A, media="receta-001")
    check("OCR + validation pass for receta-001", s == "ASK_ZONE" and "Receta validada" in r)
    s, r, _ = chat(A, "2")
    check("zona centro offers the split option", s == "ASK_MODE" and "Económicas Demo Centro + Medicity Demo Centro" in r)
    s, r, _ = chat(A, "retiro")
    check("pickup asks which option", s == "ASK_PICKUP_OPTION")
    s, r, _ = chat(A, "1")
    check("brand prompt for Paracetamol", s == "ASK_BRAND" and "Paracetamol" in r)
    s, r, _ = chat(A, "1")
    check("brand prompt for Amoxicilina", s == "ASK_BRAND" and "Amoxicilina" in r)
    s, r, _ = chat(A, "1")
    check("brand prompt for Loratadina", s == "ASK_BRAND" and "Loratadina" in r)
    s, r, _ = chat(A, "1")
    check("cart summary with total", s == "CONFIRM_CART" and "Total a pagar" in r)
    s, r, replies = chat(A, "sí")
    check("order reserved and checkout link sent", s == "AWAIT_PAYMENT" and any(m["type"] == "link" and "/checkout/" in m["link"] for m in replies))
    oid = order_from(replies)

    st, res = call("GET", f"/orders/{oid}")
    o = res["order"]
    check("order is PENDING with 3 items and active reservation", st == 200 and o["status"] == "PENDING" and len(o["items"]) == 3 and o["reservation"]["active"])
    par = next(i for i in o["items"] if i["medicine"].startswith("Paracetamol"))
    amx = next(i for i in o["items"] if i["medicine"].startswith("Amoxicilina"))
    lor = next(i for i in o["items"] if i["medicine"].startswith("Loratadina"))
    check("OTC can increase, Rx cannot", par["canIncrease"] and not amx["canIncrease"] and amx["canDecrease"])
    st, res = call("PATCH", f"/orders/{oid}/items/{par['id']}", {"quantity": par["quantity"] + 5})
    check("OTC +5 accepted and totals recalculated", st == 200 and res["order"]["subtotal"] > o["subtotal"])
    st, res = call("PATCH", f"/orders/{oid}/items/{amx['id']}", {"quantity": amx["quantity"] + 1})
    check("Rx +1 rejected (RX_INCREASE_NOT_ALLOWED)", st == 409 and res["error"]["code"] == "RX_INCREASE_NOT_ALLOWED")
    st, res = call("PATCH", f"/orders/{oid}/items/{par['id']}", {"quantity": 10000})
    check("increase beyond stock rejected (INSUFFICIENT_STOCK)", st == 409 and res["error"]["code"] == "INSUFFICIENT_STOCK")
    st, res = call("PATCH", f"/orders/{oid}/items/{lor['id']}", {"quantity": 0})
    check("quantity 0 removes the line", st == 200 and len(res["order"]["items"]) == 2)
    st, res = call("GET", f"/orders/{oid}/payment-options")
    check("payment options: card + DeUna + checkout link", st == 200 and [x["method"] for x in res["options"]] == ["card", "deuna"] and res["checkoutUrl"].endswith(oid))

    st, res = call("POST", "/payments", {"orderId": oid, "method": "deuna"})
    pay = res["payment"]
    check("DeUna intent carries the simulator link", st == 201 and pay["status"] == "PENDING" and (pay["link"] or "").startswith(WEB + "/deuna/"))
    st, res = call("GET", "/payments/banks")
    bank = res["banks"][0]["id"]
    st, res = call("POST", f"/payments/{pay['id']}/confirm", {"outcome": "rejected", "bankId": bank})
    check("DeUna rejection recorded on payment and order", st == 200 and res["payment"]["status"] == "REJECTED" and res["order"]["payment"]["status"] == "REJECTED")
    check("client told the payment was rejected", said(A, "rechazado"))
    res = pay_card(oid)
    paid = res["order"]
    check("card approval → order PAID with 2 pharmacy fulfilments (ETA 20 min)", paid["status"] == "PAID" and len(paid["fulfillments"]) == 2 and all(f["etaMinutes"] == 20 for f in paid["fulfillments"]))
    st, again = call("POST", f"/payments/{res['payment']['id']}/confirm", {"outcome": "approved"})
    check("repeated confirmation is idempotent", st == 200 and again["order"]["paidAt"] == paid["paidAt"])
    st, res = call("GET", f"/notifications?orderId={oid}&channel=pharmacy")
    check("both pharmacies notified", st == 200 and len(res["notifications"]) == 2)
    check("client got the confirmation with ETA", said(A, "Pago aprobado") and said(A, "20 minutos"))

    st, res = call("GET", "/fulfillment/orders?pharmacyId=eco-centro")
    check("cashier board lists the order", st == 200 and any(x["id"] == oid for x in res["orders"]))
    st, res = call("POST", f"/fulfillment/orders/{oid}/pharmacies/eco-centro/status", {"status": "PREPARING"})
    check("eco-centro preparing → order PREPARING", st == 200 and res["order"]["status"] == "PREPARING")
    st, res = call("POST", f"/fulfillment/orders/{oid}/pharmacies/eco-centro/status", {"status": "READY"})
    check("one pharmacy ready → order still PREPARING", st == 200 and res["order"]["status"] == "PREPARING")
    st, res = call("POST", f"/fulfillment/orders/{oid}/pharmacies/med-centro/status", {"status": "READY"})
    check("both ready → order READY", st == 200 and res["order"]["status"] == "READY")
    check("client told per pharmacy it is ready", said(A, "listo para recoger"))
    s, r, _ = chat(A, "estado")
    check("'estado' after payment reports status and completes", s == "COMPLETED" and "pagado" in r)
    s, r, _ = chat(A, "hola")
    check("next purchase asks for the cédula again", s == "ASK_ID" and "cédula" in r)


def delivery_flow():
    print("\n[2] Delivery flow, Guayaquil (no pickup option, company stock)")
    chat(B, "hola")
    chat(B, "0912345678")
    s, r, _ = chat(B, media="receta-002")
    check("receta-002 validated", s == "ASK_ZONE")
    s, r, _ = chat(B, "guayaquil")
    check("zone matched by text; pickup lacks Omeprazol but delivery works", s == "ASK_MODE" and "Omeprazol" in r and "A domicilio sí podemos" in r)
    s, r, _ = chat(B, "retiro")
    check("pickup refused in this zone", s == "ASK_MODE" and "No hay opciones de retiro" in r)
    s, r, _ = chat(B, "domicilio")
    check("asks for the address", s == "ASK_ADDRESS")
    s, r, _ = chat(B, "Av. Demo 123 y Calle 4")
    check("courier assigned + shipping fee shown", s == "ASK_BRAND" and "Repartidor" in r and "USD 2,50" in r)
    chat(B, "1")
    chat(B, "1")
    s, r, _ = chat(B, "1")
    check("delivery cart includes shipping", s == "CONFIRM_CART" and "Envío" in r)
    s, r, replies = chat(B, "si")
    check("delivery order reserved", s == "AWAIT_PAYMENT")
    oid = order_from(replies)
    st, res = call("GET", f"/orders/{oid}")
    o = res["order"]
    check("delivery order: courier, fee 2.50, total = subtotal + fee", o["delivery"]["status"] == "PENDING" and o["deliveryFee"] == 2.5 and abs(o["total"] - (o["subtotal"] + 2.5)) < 1e-9)
    res = pay_card(oid)
    check("paid delivery → CONSOLIDATING with 45 min ETA", res["order"]["delivery"]["status"] == "CONSOLIDATING" and res["order"]["delivery"]["etaMinutes"] == 45)
    check("client told the arrival ETA", said(B, "45 minutos"))
    st, res = call("GET", "/fulfillment/orders?role=courier")
    check("courier board lists the delivery", st == 200 and any(x["id"] == oid for x in res["orders"]))
    st, res = call("POST", f"/fulfillment/orders/{oid}/delivery/status", {"status": "DISPATCHED"})
    check("dispatched → order DISPATCHED, 15 min ETA", st == 200 and res["order"]["status"] == "DISPATCHED" and res["order"]["delivery"]["etaMinutes"] == 15)
    st, res = call("POST", f"/fulfillment/orders/{oid}/delivery/status", {"status": "DELIVERED"})
    check("delivered → order DELIVERED", st == 200 and res["order"]["status"] == "DELIVERED")
    check("client told 'en reparto' and 'entregado'", said(B, "en reparto") and said(B, "fue entregado"))
    st, res = call("GET", f"/notifications?orderId={oid}&channel=courier")
    check("courier notified", st == 200 and len(res["notifications"]) >= 2)


def expiry_flow():
    print("\n[3] New client, expiry, renew and cancel")
    s, r, _ = chat(C, "buenas")
    s, r, _ = chat(C, "1101234567")
    check("unknown cédula → asks for the name", s == "ASK_NAME")
    s, r, _ = chat(C, "Carla Demo")
    check("registers the new client", s == "ASK_PRESCRIPTION" and "Carla" in r)
    chat(C, media="receta-001")
    s, r, _ = chat(C, "quito norte")
    check("zona norte: single store covers everything", s == "ASK_MODE" and "toda la receta" in r)
    chat(C, "retiro")
    chat(C, "1")
    chat(C, "2")
    chat(C, "1")
    chat(C, "1")
    s, r, replies = chat(C, "sí")
    oid = order_from(replies)
    time.sleep(5.5)  # reservation TTL is 4 s in this run, worker ticks every 0.5 s
    st, res = call("GET", f"/orders/{oid}")
    check("reservation expired by the worker", res["order"]["status"] == "EXPIRED" and not res["order"]["reservation"]["active"])
    check("client told the reservation expired", said(C, "venció"))
    st, res = call("POST", "/payments", {"orderId": oid, "method": "card"})
    check("expired order cannot be paid", st == 409 and res["error"]["code"] == "RESERVATION_EXPIRED")
    s, r, _ = chat(C, "continuar")
    check("'continuar' renews the reservation and resends the link", s == "AWAIT_PAYMENT" and "nueva reserva" in r)
    st, res = call("GET", f"/orders/{oid}")
    check("order PENDING again", res["order"]["status"] == "PENDING" and res["order"]["reservation"]["active"])
    s, r, _ = chat(C, "cancelar")
    st, res = call("GET", f"/orders/{oid}")
    check("'cancelar' cancels and releases", s == "COMPLETED" and res["order"]["status"] == "CANCELLED" and said(C, "cancelado"))


def main():
    env = dict(os.environ, PORT=str(PORT), WEB_BASE_URL=WEB, STATIC_DIR="none", RESERVATION_TTL="4s", EXPIRY_TICK="500ms")
    env["PATH"] = os.path.expanduser("~/bin") + ":" + env.get("PATH", "")
    tmp = pathlib.Path(tempfile.mkdtemp(prefix="farmi-smoke-"))
    binary = tmp / "farmi"
    subprocess.run(["go", "build", "-o", str(binary), "./cmd/server"], cwd=ROOT / "backend", env=env, check=True)
    log = open(tmp / "server.log", "w")
    srv = subprocess.Popen([str(binary)], cwd=ROOT / "backend", env=env, stdout=log, stderr=subprocess.STDOUT)
    try:
        for _ in range(60):
            try:
                st, res = call("GET", "/health")
                if st == 200:
                    break
            except Exception:
                pass
            time.sleep(0.25)
        else:
            print("backend did not start; see", tmp / "server.log")
            sys.exit(2)
        print("[0] Health:", res)
        pickup_flow()
        delivery_flow()
        expiry_flow()
    finally:
        srv.terminate()
        try:
            srv.wait(timeout=10)
        except subprocess.TimeoutExpired:
            srv.kill()
    failed = [n for n, ok in checks if not ok]
    print(f"\n{len(checks) - len(failed)}/{len(checks)} checks passed" + (f"; FAILED: {failed}" if failed else ""))
    print("server log:", tmp / "server.log")
    sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
