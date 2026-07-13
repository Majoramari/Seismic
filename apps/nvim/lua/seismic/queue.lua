local M = {}

local MAX_SIZE = 100
local MAX_ATTEMPTS = 3

local queue = {}

function M.enqueue(payload)
	if #queue >= MAX_SIZE then
		table.remove(queue, 1)
	end
	table.insert(queue, { payload = payload, attempts = 0 })
end

local function try_send(payload, api_key, api_url, callback)
	local body = vim.json.encode(payload)
	vim.system({
		"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
		"-X", "POST",
		"-H", "Content-Type: application/json",
		"-H", "Authorization: Bearer " .. api_key,
		"-d", body,
		api_url .. "/api/heartbeat",
	}, { text = true }, function(result)
		local ok = result.code == 0 and result.stdout:match("^2%d%d") ~= nil
		callback(ok)
	end)
end

-- Sends only the OLDEST queued item, not the whole backlog. Sending
-- everything at once (the old behavior) guaranteed a burst of requests
-- that blew straight through the server's per-key rate limit whenever
-- more than one heartbeat was queued up — every failed retry (including
-- the resulting 429s) got requeued, so it repeated forever. One item per
-- flush call keeps retries paced naturally alongside real heartbeats.
function M.flush(api_key, api_url)
	if #queue == 0 or api_key == "" then
		return
	end

	local item = table.remove(queue, 1)

	try_send(item.payload, api_key, api_url, function(ok)
		if ok then
			return -- drained successfully, nothing to do
		end

		item.attempts = item.attempts + 1
		if item.attempts < MAX_ATTEMPTS then
			table.insert(queue, 1, item) -- put it back at the front to retry next time
		end
		-- else: dropped after MAX_ATTEMPTS failed tries
	end)
end

return M
