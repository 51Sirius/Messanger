import random

from flask import Flask, render_template, request, redirect, url_for, abort, jsonify
from flask_login import LoginManager, login_user, logout_user, current_user
from Models.User import User, db, Message, UserMessages
from flask_migrate import Migrate
from datetime import date
from Utils.forms import *
from cfg import DATABASE_URL, S_KEY

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = DATABASE_URL
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
app.config['SECRET_KEY'] = S_KEY
app.config['MAX_CONTENT_LENGTH'] = 16 * 1024 * 1024  # Max file size = 16 megaByte
db.init_app(app)
login = LoginManager(app)
login.init_app(app)
migrate = Migrate(app, db)


@login.user_loader
def load_user(id):
    return User.query.get(int(id))


@app.route('/logout')
def logout():
    logout_user()
    return redirect(url_for('index'))


@app.route('/')
def index():
    return render_template('index.html', title='MainPage')


@app.errorhandler(403)
def not_found_error(error):
    return render_template('errors/403.html'), 403


@app.errorhandler(404)
def not_found_error(error):
    return render_template('errors/404.html'), 403


@app.errorhandler(500)
def not_found_error(error):
    print(error)
    return render_template('errors/500.html'), 500

@app.route('/update/<string:username>', methods=['GET'])
def update(username):
    if not current_user.is_authenticated or username is None:
        return redirect(url_for('index'))
    
    companion = User.query.filter_by(username=username).first()
    messages_to = UserMessages.query.filter_by(id_sender=current_user.id).filter_by(id_recipient=companion.id).all()
    messages_from = UserMessages.query.filter_by(id_sender=companion.id).filter_by(id_recipient=current_user.id).all()
    messages = messages_to + messages_from
    messages = sorted(messages, key=lambda x: x.message.date_create, reverse=True)
    data = []

    for message in messages:
        status = request.get("").json()["status"]
        data.append({"id":str(message.message.id),"status":str(1 if status else 0)})
        
    return jsonify(data)


@app.route('/login', methods=['GET', 'POST'])
def login():
    if current_user.is_authenticated:
        return redirect(url_for('index'))
    forms = LoginForm()
    if forms.validate_on_submit():
        nick = forms.username.data
        user = User.query.filter_by(username=nick).first()
        if not user:
            abort(403)
        login_user(user, remember=forms.remember_me)
        return redirect(url_for('index'))
    return render_template('login.html', form=forms)


@app.route('/registration', methods=['GET', 'POST'])
def registration():
    if current_user.is_authenticated:
        return redirect(url_for('index'))
    register_form = Registration()
    if register_form.validate_on_submit():
        name = register_form.username.data
        existing_user = User.query.filter_by(username=name).first()
        if existing_user:
            abort(400)
        wallet = "2" #request.get("")
        user = User(username=name, wallet=wallet)
        db.session.add(user)
        db.session.commit()
        user = User.query.filter_by(username=name).first()
        login_user(user, remember=True)
        return redirect(url_for('index'))
    return render_template('registration.html', form=register_form)


@app.route('/search', methods=['GET', 'POST'])
def search():
    search_form = GetMessages()
    return render_template(
            'search.html',
            search_form=search_form)

@app.route('/chat/', defaults={'username': ''}, methods=['GET', 'POST'])
@app.route('/chat/<string:username>', methods=['GET', 'POST'])
def chat(username):
    if not current_user.is_authenticated:
        return redirect(url_for('index'))
    messages_to = messages_from = []
    username = request.form.get('name') if username == '' or username is None else username
    send_form = SendMessage()

    if username is None:
        return redirect(url_for('search'))
    companion = User.query.filter_by(username=username).first()

    if companion is None:
        abort(404)

    if send_form.validate_on_submit():
        data = date.today().strftime("%Y/%m/%d/%I/%M/%S")
        message = Message(text=send_form.body.data, date_create=data)
        db.session.add(message)
        db.session.commit()
        message = Message.query.all()[-1]
        user_messages = UserMessages(id_sender=current_user.id, id_recipient=companion.id, id_message=message.id)
        db.session.add(user_messages)
        db.session.commit()
        # request.post()

    messages_to = UserMessages.query.filter_by(id_sender=current_user.id).filter_by(id_recipient=companion.id).all()
    messages_from = UserMessages.query.filter_by(id_sender=companion.id).filter_by(id_recipient=current_user.id).all()
    messages = messages_to + messages_from
    messages = sorted(messages, key=lambda x: x.message.date_create, reverse=True)
    count = len(messages)
    return render_template(
        'chat.html',
        send_form=send_form,
        messages=messages,
        count=count,
        name=username)
    

if __name__ == '__main__':
    app.run()