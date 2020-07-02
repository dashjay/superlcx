local utils = require("utils")

global_login_collection = {}

function table.kIn(tbl, key)
    if tbl == nil then
        return false
    end
    for k, v in pairs(tbl) do
        if k == key then
            return true
        end
    end
    return false
end

function table.removeKey(tab, val)
    for i, v in ipairs (tab) do
        if (v.id == val) then
          tab[i] = nil
        end
    end
end

function on_http_request(req_tab)
    print("[lua] invoke lua on_http_request")
    if (not req_tab["method"] == "POST" or string.find(req_tab["url"],"login") == nil)
    then
        return false
    end
    if (req_tab["body"] == "" or req_tab["body"] == nil)
    then
        return false
    end

    local res = utils.decode(req_tab["body"])

    if (res["username"] == "" or res["password"] == "")
    then
        return false
    end
    print(string.format("[lua] login with username %s, password %s", res["username"], res["password"]))
    global_login_collection[req_tab["unique_id"]] = res
    return true
end

function on_http_response(resp_tab)
    print("[lua] invoke lua on_http_response")
    if not table.kIn(global_login_collection, resp_tab["unique_id"])
    then
        return
    end

    local obj = global_login_collection[resp_tab["unique_id"]];

    if (resp_tab["body"] == "" or resp_tab["body"] == nil)
    then
        return
    end
    local body = utils.decode(resp_tab["body"])

    local risk_level = 'medium'
    local login_success = false
    if (body["code"] == 0 or body["code"] == "0")
    then
        login_success = true
        risk_level = 'high'
    end
    table.removeKey(global_login_collection, resp_tab["unique_id"])
    print(string.format("[lua] delete key unique_id %s",resp_tab["unique_id"]))
end
