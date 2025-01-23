document.getElementById("button-update").addEventListener("click", function() {
    const uri = 'YOUR_URI_HERE';

    fetch(uri)
        .then(response => response.json())
        .then(data => {
            const statusMap = {};
            data.forEach(item => {
                statusMap[item.id] = item.status;
            });

            const messages = document.querySelectorAll('.message');

            messages.forEach(message => {
                const messageId = message.getAttribute('data-id');
                const status = statusMap[messageId];

                if (status === 1) {
                    message.classList.add('green');
                    message.classList.remove('red');
                } else if (status === 0) {
                    message.classList.add('red');
                    message.classList.remove('green');
                }
            });
        })
        .catch(error => console.error('Ошибка при получении данных:', error));
});