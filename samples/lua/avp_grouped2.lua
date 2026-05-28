--[[
This is a sample Lua script that demonstrates how to create a Diameter grouped AVP.

To run this script execute: tgdp @avp_group2.lua
]]

local d = require("diameter")
local p = require("print_avp")

-- main
--

-- Create a new AVP with the name "Vendor-Specific-Application-Id"
--
avp_vnd_spec_app_id = d.avp.new("Vendor-Specific-Application-Id", 260, d.AVP_FLAG_MANDATORY, 0, d.AVP_TYPE_GROUPED)


-- Retrive members instances from AVP store and add to the grouped AVP
--
for key, m in ipairs(avp_vnd_spec_app_id.members) do
    local value = d.avp.fetch(m)
    if value then
        err = avp_vnd_spec_app_id:set_value(value)
        if err then
            print(string.format("Failed to set value for %s: %s\n", m, err))
        end
    end
end

-- Print the grouped AVP
--
p.print_avp_data(avp_vnd_spec_app_id, 0)
