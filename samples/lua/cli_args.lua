--[[
This is a sample Lua script that demonstrates how to use command line arguments.

To run this script execute: tgdp @cli_args.lua <arg> [<arg> ...]
]]

print("Program: " .. arg[0])
print"Args:"
for i, arg in ipairs(arg) do
	print("  " .. i .. ": " .. arg)
end
