async function getOrder() {
    const orderId = document.getElementById('orderId').value;
    if (!orderId) {
        alert('Please enter Order ID');
        return;
    }

    try {
        const response = await fetch(`http://localhost:8082/order?order_uid=${orderId}`);
        if (!response.ok) {
            throw new Error(await response.text());
        }
        const order = await response.json();
        document.getElementById('result').innerHTML = syntaxHighlight(JSON.stringify(order, null, 2));
    } catch (error) {
        document.getElementById('result').innerHTML = `Error: ${error.message}`;
    }
}

async function updateOrderCount() {
    try {
        const response = await fetch('http://localhost:8082/orders');
        if (response.ok) {
            const orders = await response.json();
            document.getElementById('orderCount').textContent = orders.length;
        }
    } catch (error) {
        console.error('Failed to update order count:', error);
    }
}

function syntaxHighlight(json) {
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(
        /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
        function (match) {
            let cls = 'number';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'key';
                } else {
                    cls = 'string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'boolean';
            } else if (/null/.test(match)) {
                cls = 'null';
            }
            return `<span class="${cls}">${match}</span>`;
        }
    );
}

// Update order count every 5 seconds
updateOrderCount();
setInterval(updateOrderCount, 5000);