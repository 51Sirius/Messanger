from flask_sqlalchemy import SQLAlchemy
from flask_login import UserMixin
import random, hashlib


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
    hash = db.Column(db.String(255), nullable=True)
    date_create = db.Column(db.String(64))

    def set_hash(self, wallet):
        m = hashlib.sha256() 
        m.update(wallet.encode('utf-8'))
        m.update(b" "+self.text.encode('utf-8'))
        tmp = m.hexdigest()
        self.hash = tmp[:200] if len(tmp) > 200 else tmp
    

class UserMessages(db.Model):
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    id_sender = db.Column(db.Integer, db.ForeignKey('user.id'))
    id_message = db.Column(db.Integer, db.ForeignKey('message.id'))
    message = db.relationship('Message', backref=db.backref('message', lazy=True))
    id_recipient = db.Column(db.Integer, db.ForeignKey('user.id'))

    def get_sender(self):
        return User.query.filter_by(id=self.id_sender).first()