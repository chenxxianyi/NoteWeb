"""Test AI mock endpoints — summary, explain, translate, chat."""


class TestSummary:
    def test_summary(self, client, auth_headers):
        # Upload a document first
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("ai.txt", b"AI test document content for summary.", "text/plain")},
        )
        doc_id = up.json()["id"]

        resp = client.get(f"/api/v1/ai/documents/{doc_id}/summary")
        assert resp.status_code == 200
        assert "summary" in resp.json()
        assert len(resp.json()["summary"]) > 0


class TestExplain:
    def test_explain(self, client):
        resp = client.post("/api/v1/ai/explain", json={"text": "人工智能正在改变世界"})
        assert resp.status_code == 200
        assert "explanation" in resp.json()
        assert len(resp.json()["explanation"]) > 0

    def test_explain_empty(self, client):
        resp = client.post("/api/v1/ai/explain", json={"text": ""})
        assert resp.status_code == 200
        assert "explanation" in resp.json()


class TestTranslate:
    def test_translate(self, client):
        resp = client.post("/api/v1/ai/translate", json={
            "text": "Hello World", "target_lang": "zh",
        })
        assert resp.status_code == 200
        assert "translation" in resp.json()


class TestChat:
    def test_chat(self, client):
        resp = client.post("/api/v1/ai/chat", json={
            "document_id": 1, "question": "这篇文章讲了什么？",
        })
        assert resp.status_code == 200
        assert "answer" in resp.json()
