--[[
This is a sample Lua script that demonstrates how to create a Diameter AVP.

To run this script execute: tgdp @avp.lua
]]

local d = require("diameter")
local p = require("print_avp")

-- main
--

-- Create a new AVP with the name "Route-Record", code 282, mandatory flag, vendor ID 0, and type 'Identity'
--
avp_route_rec = d.avp.new("Route-Record", 282, d.AVP_FLAG_MANDATORY, 0, d.AVP_TYPE_IDENTITY)

-- Set the value of the AVP using the 'set_value' method
--
err = avp_route_rec:set_value("example.mnc000.mcc999.3gppnetwork.org")
if err ~= nil then
  print(err)
end
p.print_avp_data(avp_route_rec, 0)

-- Change the value of the AVP using 'value' property
--
avp_route_rec.value = "example.epc.mnc001.mcc999.3gppnetwork.org"
print(string.format("%s (%d): %s", avp_route_rec.name, avp_route_rec.code, avp_route_rec.value))

-- Create a new AVP with the name "User-Name" getting from 'Pkl' template
--
avp_user_name, err = d.avp.fetch("User-Name")[1]
if err ~= nil then
  print(err)
end
avp_user_name.value = "1234501234567890"
p.print_avp_data(avp_user_name, 0)

-- Fetch an existing AVP instance[s] from the AVP store
-- Note: AVP store is the part of module "diameter"
--
avp_aut_app_id, err = d.avp.fetch("Auth-Application-Id")
if err ~= nil then
  print(err)
end
for key,avp in ipairs(avp_aut_app_id) do
    p.print_avp_data(avp, 0)
end
