--[[
This is a sample Lua script that demonstrates how to create a Diameter message by getting from template
defined in Pkl configuration file.
AVPs data loading from 'avp-data.yaml'.

To run this script execute: tgdp @message2.lua
]]

local d = require("diameter")
local print = io.write

function print_avp_data(avp, shift)
    for i=1,shift,1 do print(" ") end

    if avp:is_grouped() then
        print(string.format("  %s (%d):\n", avp.name, avp.code))
        for _,member in ipairs(avp.members) do
            print_avp_data(member, 2)
        end
    else
        print(string.format("  %s (%d): %s\n", avp.name, avp.code, avp.value))
    end
end

ulr, err = d.message.get("S6a", "UL", true)
if err ~= nil then
    print(err)
    return 1
end

print(string.format("App ID: %d\n", ulr.app_id))
print(string.format("Cmd Code: %d\n", ulr.cmd_code))
print(string.format("Flags: %X\n", ulr.flags))
print(string.format("Hop by Hop: 0x%X\n", ulr.hop_by_hop))
print(string.format("End to End: 0x%X\n", ulr.end_to_end))

print("AVPs:\n")
for key,avp in ipairs(ulr.avps) do
    print_avp_data(avp, 0)
end
