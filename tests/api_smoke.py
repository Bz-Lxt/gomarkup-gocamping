#!/usr/bin/env python3
"""Mock-mode API smoke. Expected cost ¥0."""
import json
import os
import sys
import urllib.request

BASE = os.environ.get("API_BASE", "http://127.0.0.1:28313")


def call(method, path, body=None, token=None):
    data = None if body is None else json.dumps(body).encode()
    req = urllib.request.Request(BASE + path, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    if token:
        req.add_header("Authorization", "Bearer " + token)
    with urllib.request.urlopen(req, timeout=20) as resp:
        return json.loads(resp.read().decode())


def main():
    h = call("GET", "/healthz")
    assert h["code"] == 0, h
    login = call("POST", "/api/v1/auth/login", {"username": "leader", "password": "leader123"})
    assert login["code"] == 0, login
    tok = login["data"]["token"]
    me = call("GET", "/api/v1/me", token=tok)
    assert me["data"]["username"] == "leader"
    books = call("GET", "/api/v1/route-books", token=tok)
    assert books["code"] == 0
    print("[PASS] Health Check")
    print("[PASS] Auth Login")
    print("[PASS] Route list")
    print("COST ¥0")
    return 0


if __name__ == "__main__":
    sys.exit(main())
