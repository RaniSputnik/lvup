-- This file will be included with love games built by LVUP for debugging
-- It is not intended to be included in builds for distribution.
-- Unanswered questions & TODOs;
-- 1. What happens in the error scenarios? Are they working correctly.
-- 2. Respond to events from the Go server, update global variables, restart the game etc.
--    (We will need to send messages from the server to the main thread, love event is not defined here)
-- 3. How to kill other threads that have been started by the game?
-- 4. Inject this script so that the user doesn't need to write it themselves.

function log(msg)
    print('CLIENT::'..msg)
end

log('Dialing localhost:8080')

socket = require('socket');
local client = socket.tcp();

success, err = client:connect("localhost", "8080")
if err ~= nil then
    log(err)
    return
end

-- success, err = client:send("Hello there")
-- if err ~= nil then
--     log(err)
--     return
-- end

local err = nil
while true do
    val, err = client:receive("*l")
    if err == 'closed' then
        log('Server closed the connection')
        return
    end
    log('Command received, '..val)

    --if val == 'restart' then
    --    love.event.quit( "restart" )
    --end
end