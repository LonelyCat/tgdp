--[[
This is a sample Lua script that demonstrates how to create a Diameter message.

To run this script execute: tgdp @message1.lua
]]

local d = require("diameter")

-- Create a new Diameter message without an AVP
ulr, err = d.message.new("S6a", "UL", true)
if err ~= nil then
    print(err)
    return 1
end

-- Add AVPs
-- Get AVPs and values from templates
local avp_sess_id, err = d.avp.fetch("Session-Id")[1]
if err ~= nil then
    print(err)
    return 1
end

local avp_dst_realm, err = d.avp.fetch("Destination-Realm")[1]
if err ~= nil then
    print(err)
    return 1
end

-- Overwrite AVP value via property 'value'
avp_sess_id.value = "my.session.example.com"

-- Overwrite AVP value via method 'set_value'
-- If needed of course :)
avp_dst_realm:set_value("example.com")

-- Create a new AVP from scratch
local avp_imsi, err = d.avp.new("User-Name", 1, d.AVP_FLAG_MANDATORY, 0, d.AVP_TYPE_UTF8_STRING)
if err ~= nil then
    print(err)
    return 1
end
avp_imsi.value = "123451234567890"

-- Add AVPs to the message
ulr:add_avp(avp_sess_id)
ulr:add_avp(avp_dst_realm)
ulr:add_avp(avp_imsi)

-- Print message header ...
print(string.format("App ID: %d", ulr.app_id))
print(string.format("Cmd Code: %d", ulr.cmd_code))
print(string.format("Flags: %X", ulr.flags))
print(string.format("Hop by Hop: 0x%X", ulr.hop_by_hop))
print(string.format("End to End: 0x%X", ulr.end_to_end))

-- .. and AVPs
print("AVPs:")
avps = ulr.avps
for key,avp in ipairs(avps) do
    print(string.format("  %s (%d): %s", avp.name, avp.code, avp:to_text()))
end
