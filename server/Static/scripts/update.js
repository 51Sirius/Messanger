document.getElementById("button-refresh").addEventListener("click", function() {
    const tmp = document.querySelector('.main').getAttribute('data-id');

    const uri = "http://127.0.0.1:5000/update/"+tmp;
    console.log(uri)
    fetch(uri)
        .then(response => response.json())
        .then(data => {
            const statusMap = {};
            data.forEach(item => {
                statusMap[item.id] = item.status;
            });
            console.log(statusMap)
            const messages = document.querySelectorAll('.msg-box');

            messages.forEach(message => {
                const messageId = message.getAttribute('data-id');
                const status = statusMap[messageId];

                if (status === "1") {
                    message.classList.add('green');
                    message.classList.remove('red');
                } else if (status === "0") {
                    message.classList.add('red');
                    message.classList.remove('green');
                }
            });
        })
        .catch(error => console.error('Ошибка при получении данных:', error));
});