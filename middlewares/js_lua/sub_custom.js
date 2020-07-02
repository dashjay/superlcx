var global_login_collection = {};

var create_login_obj = function (username, password) {
    return {"username": username, "password": password};
};

var on_http_request = function (req_tab) {
    console.log("[js] invoke js on_http_request")
    if (req_tab["method"] !== "POST" || req_tab["url"].indexOf("login") === -1) {
        return false;
    }
    if (req_tab["body"] === "") {
        return false;
    }
    var body = JSON.parse(req_tab["body"]);
    if (body["username"] === "" || body["password"] === "") {
        return false;
    }
    console.log("[js] login with username:", body["username"], "password: ", body["password"])
    global_login_collection[req_tab["unique_id"]] = create_login_obj(body["username"], body["password"]);
    return true;
};

var on_http_response = function (resp_tab) {
    console.log("[js] invoke js on_http_response")
    if (!global_login_collection.hasOwnProperty(resp_tab["unique_id"])) {
        return;
    }
    var obj = global_login_collection[resp_tab["unique_id"]];

    delete global_login_collection[resp_tab["unique_id"]];

    console.log("[js] delete key unique_id", resp_tab["unique_id"])
};

