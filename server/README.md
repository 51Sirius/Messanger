### Instalation

```
cd server
```

```
python -m venv .venv
.venv/Scripts/activate
pip install -r requirements.txt
flask db init
```

```
python
>>> from Models.User import db
>>> from app import app
>>> app.app = app
>>> db.app = app
>>> app.app_context().push()
>>> db.create_all()
```