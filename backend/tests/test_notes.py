"""Test note endpoints — CRUD operations."""


class TestCreate:
    def test_create_note(self, client, auth_headers):
        resp = client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "My Note", "content": "# Hello\nWorld", "tags": ["test"]},
        )
        assert resp.status_code == 200
        data = resp.json()
        assert data["title"] == "My Note"
        assert data["content"] == "# Hello\nWorld"
        assert data["tags"] == ["test"]

    def test_create_note_with_document(self, client, auth_headers):
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("ref.txt", b"ref content", "text/plain")},
        )
        doc_id = up.json()["id"]

        resp = client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "Ref Note", "content": "from doc", "document_id": doc_id},
        )
        assert resp.status_code == 200
        assert resp.json()["document_id"] == doc_id
        assert resp.json()["document_title"] is not None

    def test_create_without_auth(self, client):
        resp = client.post("/api/v1/notes", json={"title": "x"})
        assert resp.status_code == 401


class TestList:
    def test_list_notes(self, client, auth_headers):
        client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "N1", "content": "C1", "tags": ["a"]},
        )
        resp = client.get("/api/v1/notes", headers=auth_headers)
        assert resp.status_code == 200
        assert len(resp.json()) >= 1

    def test_list_filter_by_document(self, client, auth_headers):
        resp = client.get("/api/v1/notes?document_id=1", headers=auth_headers)
        assert resp.status_code == 200

    def test_list_filter_by_tag(self, client, auth_headers):
        client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "tagged", "content": "...", "tags": ["python"]},
        )
        resp = client.get("/api/v1/notes?tag=python", headers=auth_headers)
        assert resp.status_code == 200

    def test_list_empty(self, client, auth_headers):
        resp = client.get("/api/v1/notes", headers=auth_headers)
        assert resp.status_code == 200
        assert isinstance(resp.json(), list)


class TestGet:
    def test_get_note(self, client, auth_headers):
        created = client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "Find Me", "content": "notes content"},
        )
        note_id = created.json()["id"]

        resp = client.get(f"/api/v1/notes/{note_id}", headers=auth_headers)
        assert resp.status_code == 200
        assert resp.json()["title"] == "Find Me"

    def test_get_note_not_found(self, client, auth_headers):
        resp = client.get("/api/v1/notes/99999", headers=auth_headers)
        assert resp.status_code == 404


class TestUpdate:
    def test_update_note_title(self, client, auth_headers):
        created = client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "Old Title", "content": "x"},
        )
        note_id = created.json()["id"]

        resp = client.patch(f"/api/v1/notes/{note_id}",
            headers=auth_headers,
            json={"title": "New Title"},
        )
        assert resp.status_code == 200
        assert resp.json()["title"] == "New Title"

    def test_update_note_content(self, client, auth_headers):
        created = client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "C", "content": "old"},
        )
        note_id = created.json()["id"]

        resp = client.patch(f"/api/v1/notes/{note_id}",
            headers=auth_headers,
            json={"content": "new content"},
        )
        assert resp.status_code == 200
        assert resp.json()["content"] == "new content"

    def test_update_not_found(self, client, auth_headers):
        resp = client.patch("/api/v1/notes/99999",
            headers=auth_headers,
            json={"title": "x"},
        )
        assert resp.status_code == 404


class TestDelete:
    def test_delete_note(self, client, auth_headers):
        created = client.post("/api/v1/notes",
            headers=auth_headers,
            json={"title": "Delete Me", "content": "bye"},
        )
        note_id = created.json()["id"]

        resp = client.delete(f"/api/v1/notes/{note_id}", headers=auth_headers)
        assert resp.status_code == 200

        # Verify deleted
        check = client.get(f"/api/v1/notes/{note_id}", headers=auth_headers)
        assert check.status_code == 404

    def test_delete_not_found(self, client, auth_headers):
        resp = client.delete("/api/v1/notes/99999", headers=auth_headers)
        assert resp.status_code == 404
