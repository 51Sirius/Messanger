from flask_sqlalchemy import SQLAlchemy
from flask_login import UserMixin


db = SQLAlchemy()

class User(db.Model, UserMixin):
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    username = db.Column(db.String(64), index=True, unique=True)
    wallet = db.Column(db.String(255), index=True, unique=True)


    def __repr__(self):
        return '<User {}>'.format(self.username)


class Message(db.Model):
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    text = db.Column(db.String(2048), nullable=True)
    date_create = db.Column(db.String(64))
    

class UserMessages(db.Model):
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    id_sender = db.Column(db.Integer, db.ForeignKey('user.id'))
    id_message = db.Column(db.Integer, db.ForeignKey('message.id'))
    message = db.relationship('Message', backref=db.backref('message', lazy=True))
    id_recipient = db.Column(db.Integer, db.ForeignKey('user.id'))