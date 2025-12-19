import http from "k6/http";
import { sleep } from "k6";
import { check } from "k6";

export const options = {
    vus: 50,          // 50 user simultaniously
    duration: "10s",  // run 10 secs
};

export default function () {
    const userID = Math.floor(Math.random() * 1000000).toString();
    const productID = "p1";

    const payload = JSON.stringify({
        user_id: userID,
        product_id: productID
    });

    const headers = { "Content-Type": "application/json" };

    const res = http.post("http://localhost:8080/flashsale/precheck", payload, { headers });

    check(res, {
        "status is 200": (r) => r.status === 200,
    });

    sleep(0.1); // slow down
}
