--[[
This is a sample Lua script that demonstrates how to create a Diameter AVP.

To run this script execute: tgdp @avp.lua
]]

local d = require("diameter")

-- Create a new AVP with the name "Route-Record", code 282, mandatory flag, vendor ID 0, and type 'Identity'
--
avp = d.avp.new("Route-Record", 282, d.AVP_FLAG_MANDATORY, 0, d.AVP_TYPE_IDENTITY)

-- Set the value of the AVP using the 'set_value' method
--
err = avp:set_value("example.mnc000.mcc999.3gppnetwork.org")
if err ~= nil then
  print(err)
end
print(string.format("%s (%d): %s", avp.name, avp.code, avp.value))

-- Change the value of the AVP using 'value' property
--
avp.value = "example.epc.mnc000.mcc999.3gppnetwork.org"
print(string.format("%s (%d): %s", avp.name, avp.code, avp.value))

-- Create a new AVP with the name "User-Name" getting from 'Pkl' template
--
avp, err = d.avp.get("User-Name")
if err ~= nil then
  print(err)
end
avp.value = "1234501234567890"
print(string.format("%s (%d): %s", avp.name, avp.code, avp.value))
