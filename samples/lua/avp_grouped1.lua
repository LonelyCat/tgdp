--[[
This is a sample Lua script that demonstrates how to create a Diameter grouped AVP.

To run this script execute: tgdp @avp_group1.lua
]]

local d = require("diameter")
local p = require("print_avp")

-- main
--

-- Create a new AVP with the name "User-Identity"
--
avp_user_identity = d.avp.new("User-Identity", 700, d.AVP_FLAG_MANDATORY+d.AVP_FLAG_VENDOR_SPECIFIC, 10415, d.AVP_TYPE_GROUPED)

-- Member of "User-Identity" AVP are:
--   Public-Identity
--   MSISDN
avp_pub_identity = d.avp.new("Public-Identity", 601, d.AVP_FLAG_MANDATORY+d.AVP_FLAG_VENDOR_SPECIFIC, 10415, d.AVP_TYPE_UTF8_STRING)
avp_msisdn = d.avp.new("MSISDN", 701, d.AVP_FLAG_MANDATORY+d.AVP_FLAG_VENDOR_SPECIFIC, 10415, d.AVP_TYPE_OCTET_STRING)

avp_pub_identity.value = "sip:+1234567890@ims.mnc888.mcc999.3gppnetwork.org"
avp_msisdn.value = "1234567890"

avp_user_identity:set_value({avp_pub_identity, avp_msisdn})

-- Print the grouped AVP
--
p.print_avp_data(avp_user_identity, 0)
