--[[
This is a sample Lua script that demonstrates how to create a Diameter peer using the Lua Diameter library.
Script creates a Diameter peer named "DEMO" with address "localhost", port 3869, and transport protocol "TCP",
connects to it, and then disconnects from it.

To run this script execute: tgdp @peer.lua
Add flag '-v 3' to increase verbosity level and see coomon messages: tgdp -v 3 @peer.lua
]]

local d = require("diameter")

local peer, err = d.peer.new("DEMO", "localhost", 3869, "tcp")
if err ~= nil then
  print(err)
  return
end

print(string.format("Peer %s created:", peer.name))
print(string.format("  address: %s", peer.address))
print(string.format("  port: %d", peer.port))
print(string.format("  transport: %s", peer.transport))

err = peer:connect()
if err ~= nil then
  print(err)
else
  print(string.format("Peer %s connected successfuly", peer.name))
end

err = peer:disconnect()
if err ~= nil then
  print(err)
else
  print(string.format("Peer %s disconnected successfuly", peer.name))
end
