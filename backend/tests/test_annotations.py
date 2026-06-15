"""Test annotation endpoints — create, list, delete."""


class TestCreate:
    def test_create_annotation(self, client, auth_headers):
        # Upload a document first
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("annot.txt", b"Test document for annotation.", "text/plain")},
        )
        doc_id = up.json()["id"]

        resp = client.post("/api/v1/annotations",
            headers=auth_headers,
            json={
                "document_id": doc_id,
                "page": 1,
                "selected_text": "Test document",
                "color": "#FFD700",
                "type": "highlight",
                "position_data": {"x": 10, "y": 20, "width": 100, "height": 20},
            },
        )
        assert resp.status_code == 200
        data = resp.json()
        assert data["document_id"] == doc_id
        assert data["selected_text"] == "Test document"
        assert data["color"] == "#FFD700"
        assert data["type"] == "highlight"

    def test_create_without_auth(self, client):
        resp = client.post("/api/v1/annotations", json={"document_id": 1})
        assert resp.status_code == 401


class TestList:
    def test_list_by_document(self, client, auth_headers):
        # Upload doc
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("list.txt", b"Content", "text/plain")},
        )
        doc_id = up.json()["id"]

        # Create annotation
        client.post("/api/v1/annotations",
            headers=auth_headers,
            json={"document_id": doc_id, "selected_text": "hello"},
        )

        resp = client.get(f"/api/v1/documents/{doc_id}/annotations", headers=auth_headers)
        assert resp.status_code == 200
        assert len(resp.json()) >= 1

    def test_list_empty(self, client, auth_headers):
        resp = client.get("/api/v1/documents/999/annotations", headers=auth_headers)
        assert resp.status_code == 200
        assert resp.json() == []


class TestDelete:
    def test_delete_annotation(self, client, auth_headers):
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("delann.txt", b"Data", "text/plain")},
        )
        doc_id = up.json()["id"]

        ann = client.post("/api/v1/annotations",
            headers=auth_headers,
            json={"document_id": doc_id, "selected_text": "del"},
        )
        ann_id = ann.json()["id"]

        resp = client.delete(f"/api/v1/annotations/{ann_id}", headers=auth_headers)
        assert resp.status_code == 200

        # Verify deleted
        ls = client.get(f"/api/v1/documents/{doc_id}/annotations", headers=auth_headers)
        assert len(ls.json()) == 0

    def test_delete_not_found(self, client, auth_headers):
        resp = client.delete("/api/v1/annotations/99999", headers=auth_headers)
        assert resp.status_code == 404
