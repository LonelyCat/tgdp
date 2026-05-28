--[[
This is a sample Lua script that demonstrates how to print a Diameter AVP using the Lua Diameter library.
]]

local M = {}

local print = io.write

function M.print_avp_data(avp, shift)
    for i=1,shift,1 do print(" ") end

    if avp:is_grouped() then
        print(string.format("%s (%d):\n", avp.name, avp.code))
        for _,member in ipairs(avp.value) do
            M.print_avp_data(member, shift + 2)
        end
    else
        print(string.format("%s (%d): %s\n", avp.name, avp.code, avp:to_text()))
    end
end

return M
