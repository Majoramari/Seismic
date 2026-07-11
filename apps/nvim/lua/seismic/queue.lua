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

function M.flush(api_key, api_url)
	if #queue == 0 or api_key == "" then
		return
	end

	local still_queued = {}
	local pending = #queue
	local remaining_items = vim.deepcopy(queue)
	queue = {}

	if pending == 0 then
		return
	end

	for _, item in ipairs(remaining_items) do
		try_send(item.payload, api_key, api_url, function(ok)
			if not ok then
				item.attempts = item.attempts + 1
				if item.attempts < MAX_ATTEMPTS then
					table.insert(still_queued, item)
				end
			end
			pending = pending - 1
			if pending == 0 then
				for _, i in ipairs(still_queued) do
					table.insert(queue, i)
				end
			end
		end)
	end
end

return M