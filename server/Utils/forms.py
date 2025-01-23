from flask_wtf import FlaskForm
from wtforms import StringField, BooleanField, SubmitField
from wtforms.validators import DataRequired
from wtforms.widgets import TextInput, TextArea


class Registration(FlaskForm):
    username = StringField('Никнэйм', validators=[DataRequired()], widget=TextInput(),
                           render_kw={"placeholder": "Никнейм"})
    register = SubmitField('Регистрация')


class LoginForm(FlaskForm):
    username = StringField('Никнэйм', validators=[DataRequired()], widget=TextInput(),
                           render_kw={"placeholder": "Никнейм"})
    remember_me = BooleanField('Запомнить меня')
    submit = SubmitField('Войти')


class SendMessage(FlaskForm):
    body = StringField('Текст сообщения', validators=[DataRequired()], widget=TextArea())
    submit = SubmitField('Отправить')


class GetMessages(FlaskForm):
    name = StringField('Логин пользователя', widget=TextInput())
    submit = SubmitField('Найти')