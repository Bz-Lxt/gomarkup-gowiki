#!/usr/bin/env python3
"""GoWiki API smoke. Cost ¥0. Run inside Docker network or against published ports."""
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request

BASE = os.environ.get("GOWIKI_API", "http://web/api/v1")


def req(method, path, data=None, token=None, expect=200):
    url = BASE + path
    if "q=" in path:
        prefix, q = path.split("q=", 1)
        url = BASE + prefix + "q=" + urllib.parse.quote(q)
    body = None if data is None else json.dumps(data).encode()
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    r = urllib.request.Request(url, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(r, timeout=20) as resp:
            raw = resp.read().decode()
            code = resp.status
            payload = json.loads(raw)
    except urllib.error.HTTPError as e:
        raw = e.read().decode()
        code = e.code
        payload = json.loads(raw) if raw else {}
    if code != expect:
        raise SystemExit(f"FAIL {method} {path} -> {code} {payload} expected {expect}")
    return payload


def main():
    health = req("GET", "/health")
    assert health["data"]["status"] == "ok"
    print("[PASS] Health Check")

    auth = req("POST", "/auth/login", {"email": "admin@gowiki.dev", "password": "admin123"})
    token = auth["data"]["accessToken"]
    print("[PASS] Login")

    spaces = req("GET", "/spaces", token=token)
    sid = spaces["data"][0]["id"]
    created = req(
        "POST",
        "/documents",
        {"spaceId": sid, "title": "冒烟文档", "editorMode": "markdown"},
        token=token,
        expect=201,
    )
    doc_id = created["data"]["id"]
    print("[PASS] Create document")

    tree = req("GET", f"/documents?spaceId={sid}", token=token)
    assert any(d["id"] == doc_id for d in tree["data"])
    print("[PASS] Tree list")

    req("POST", f"/documents/{doc_id}/move", {"parentId": doc_id, "sortOrder": 0}, token=token, expect=409)
    print("[PASS] Cycle rejected")

    req("PATCH", f"/documents/{doc_id}", {"contentMd": "协同检索版本"}, token=token)
    ver = req("POST", f"/documents/{doc_id}/versions", {"label": "冒烟版本"}, token=token, expect=201)
    vid = ver["data"]["id"]
    req("GET", f"/versions/diff?left={vid}&right=current", token=token)
    print("[PASS] Version + Diff")

    req("POST", f"/versions/{vid}/rollback", token=token)
    print("[PASS] Rollback")

    req("GET", "/search?q=协同", token=token)
    print("[PASS] Search")

    req("GET", "/workbench", token=token)
    print("[PASS] Workbench")
    print("[PASS] API smoke complete")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print("FAIL", e)
        sys.exit(1)
