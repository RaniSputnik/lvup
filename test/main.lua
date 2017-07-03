thread = love.thread.newThread('net.lua')
thread:start()

local xx = 0

function love.draw()
    love.graphics.print('Hello world', xx,200)
    xx = xx + 1

    local err = thread:getError()
    if err then
        print(err)
    end
end