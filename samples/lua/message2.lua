--[[
This is a sample Lua script that demonstrates how to create a Diameter message by getting from template
defined in Pkl configuration file.
AVPs data loading from 'avp-data.yaml'.

To run this script execute: tgdp @message2.lua
]]

local d = require("diameter")
local p = require("print_avp")

-- main
--
ulr, err = d.message.fetch("S6a", "UL", true)
if err ~= nil then
    print(err)
    return 1
end

print(string.format("App ID: %d", ulr.app_id))
print(string.format("Cmd Code: %d", ulr.cmd_code))
print(string.format("Flags: %X", ulr.flags))
print(string.format("Hop by Hop: 0x%X", ulr.hop_by_hop))
print(string.format("End to End: 0x%X", ulr.end_to_end))

print("AVPs:")
for _,avp in ipairs(ulr.avps) do
    p.print_avp_data(avp, 2)
end
