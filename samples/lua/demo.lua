--[[
This is a sample Lua script that demonstrates TGDP's Lua API.

To run this script execute: tgdp @demo.lua <IMSI>
]]

local d = require("diameter")

if #arg < 1 then
    print(string.format("Usage: %s <IMSI>", arg[0]))
    return
end

local imsi = arg[1]

-- Global parameters
--
-- Peer
local peer_addr = '172.19.101.201'
local peer_port = 3868
local peer_proto = 'sctp'

-- Origing host
--
local org_realm = 'vn.mnc040.mcc250.3gppnetwork.org'
local org_host = 'hss01.vn.mnc040.mcc250.3gppnetwork.org'
local sess_id = org_host

-- Destination host
--
local dest_realm = 'vn.mnc040.mcc250.3gppnetwork.org'
local dest_host = 'mme01.vn.mnc040.mcc250.3gppnetwork.org'

--
--
local dra, err = d.peer.new("DRA", peer_addr, peer_port, peer_proto)
if err ~= nil then
    print(err)
    return
end
d.dump(dra)

err = dra:connect()
if err ~= nil then
    print(err)
    return
end
print("Connected to " .. dra.name)

idr, err = d.message.get(16777251, "ID", true) -- Insert-Subscriber-Data request
if err ~= nil then
    print(err)
    return 1
end


idr:set_avp_value("Session-Id", sess_id)
idr:set_avp_value("Origin-Realm", org_realm)
idr:set_avp_value("Origin-Host", org_host)
idr:set_avp_value("Destination-Realm", dest_realm)
idr:set_avp_value("Destination-Host", dest_host)
idr:set_avp_value("User-Name", imsi)
idr:set_avp_value("IDR-Flags", 8) -- request EPS-Location-Information
-- Dump request message
--
d.dump(idr)

-- Send request to peer
--
err = dra:send_to(idr)
if err ~= nil then
    print(err)
    dra:disconnect()
    return
end

-- Wait for response
--
local ida, err = dra:recv_from()
if err ~= nil then
    print(err)
    dra:disconnect()
    return
end

-- Check resultcode
--
local res_code, err = ida:get_avp_value("Result-Code")
if err ~= nil then
    print(err)
    dra:disconnect()
    return
end

if res_code ~= 2001 then
    print("Request failed: ", res_code)
    dra:disconnect()
    return
end

-- Dump answer message
--
d.dump(ida)

local err = dra:disconnect()
if err ~= nil then
    print(err)
end
print("Disconnected from " .. dra.name)
