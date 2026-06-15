"""Test document endpoints — upload, list, detail, rename, delete."""


class TestUpload:
    def test_upload_txt_success(self, client, auth_headers):
        resp = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("test.txt", b"Hello World\nTest content.", "text/plain")},
        )
        assert resp.status_code == 200
        data = resp.json()
        assert data["title"] == "test"
        assert data["file_type"] == "txt"
        assert data["file_size"] > 0
        assert data["id"] > 0

    def test_upload_md_file(self, client, auth_headers):
        resp = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("readme.md", b"# Title\nSome markdown.", "text/markdown")},
        )
        assert resp.status_code == 200
        assert resp.json()["file_type"] == "md"

    def test_upload_without_auth(self, client):
        resp = client.post("/api/v1/documents/upload",
            files={"file": ("test.txt", b"data", "text/plain")},
        )
        assert resp.status_code == 401

    def test_upload_multiple_files(self, client, auth_headers):
        # Upload two files
        client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("file1.txt", b"Content 1", "text/plain")},
        )
        resp = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("file2.txt", b"Content 2", "text/plain")},
        )
        assert resp.status_code == 200


class TestList:
    def test_list_empty(self, client, auth_headers):
        resp = client.get("/api/v1/documents", headers=auth_headers)
        assert resp.status_code == 200
        assert isinstance(resp.json(), list)

    def test_list_with_files(self, client, auth_headers):
        # Upload a file first
        client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("doc.txt", b"Content", "text/plain")},
        )
        resp = client.get("/api/v1/documents", headers=auth_headers)
        assert resp.status_code == 200
        data = resp.json()
        assert len(data) >= 1
        assert data[0]["title"] == "doc"

    def test_list_filter_by_type(self, client, auth_headers):
        resp = client.get("/api/v1/documents?type=txt", headers=auth_headers)
        assert resp.status_code == 200


class TestDetail:
    def test_get_detail(self, client, auth_headers):
        # Upload
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("detail.txt", b"Detail content.", "text/plain")},
        )
        doc_id = up.json()["id"]

        resp = client.get(f"/api/v1/documents/{doc_id}", headers=auth_headers)
        assert resp.status_code == 200
        data = resp.json()
        assert data["id"] == doc_id
        assert "parsed_content" in data

    def test_get_detail_not_found(self, client, auth_headers):
        resp = client.get("/api/v1/documents/99999", headers=auth_headers)
        assert resp.status_code == 404

    def test_get_detail_other_user(self, client, auth_headers):
        # Upload with tester
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("mine.txt", b"secret", "text/plain")},
        )
        doc_id = up.json()["id"]

        # Register another user
        resp = client.post("/api/v1/auth/register", json={
            "username": "other", "email": "other@test.com",
            "password": "Abcd1234", "confirm_password": "Abcd1234",
        })
        other_token = resp.json()["token"]

        # Try to access with other user
        resp = client.get(f"/api/v1/documents/{doc_id}", headers={
            "Authorization": f"Bearer {other_token}",
        })
        assert resp.status_code == 404


class TestRename:
    def test_rename_success(self, client, auth_headers):
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("old.txt", b"data", "text/plain")},
        )
        doc_id = up.json()["id"]

        resp = client.patch(f"/api/v1/documents/{doc_id}",
            headers=auth_headers,
            json={"title": "new-name"},
        )
        assert resp.status_code == 200
        assert resp.json()["title"] == "new-name"

    def test_rename_not_found(self, client, auth_headers):
        resp = client.patch("/api/v1/documents/99999",
            headers=auth_headers,
            json={"title": "x"},
        )
        assert resp.status_code == 404


class TestDelete:
    def test_delete_success(self, client, auth_headers):
        up = client.post("/api/v1/documents/upload",
            headers=auth_headers,
            files={"file": ("delete.txt", b"data", "text/plain")},
        )
        doc_id = up.json()["id"]

        resp = client.delete(f"/api/v1/documents/{doc_id}", headers=auth_headers)
        assert resp.status_code == 200

        # Should not appear in list after deletion
        ls = client.get("/api/v1/documents", headers=auth_headers)
        ids = [d["id"] for d in ls.json()]
        assert doc_id not in ids

    def test_delete_not_found(self, client, auth_headers):
        resp = client.delete("/api/v1/documents/99999", headers=auth_headers)
        assert resp.status_code == 404
