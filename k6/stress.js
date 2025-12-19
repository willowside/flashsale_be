import http from 'k6/http';

export let options = {
    stages: [
        { duration: "5s", target: 200 },
        { duration: "15s", target: 200 },
        { duration: "5s", target: 0 },
    ],
};

export default function () {
    const url = 'http://localhost:8080/flashsale/precheck';

    const uid = Math.floor(Math.random() * 1000000);

    const payload = JSON.stringify({
        user_id: `user_${uid}`,
        product_id: '1001',
    });

    http.post(url, payload, {
        headers: { 
            "Content-Type": "application/json",
            "X-User-ID": `user_${uid}`
         }
    });
}
