```
python -m venv .venv
.venv/Scripts/activate

flask db init
flask db migratge -m "Init migrate"
```


```
python
>>> from Models.User import db
>>> from app import app
>>> app.app = app
>>> db.app = app
>>> db.init_app(app)
>>> app.app_context().push()
>>> db.create_all()
```