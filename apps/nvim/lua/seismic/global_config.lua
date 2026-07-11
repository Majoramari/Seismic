local M = {}

local path = vim.fn.expand("~/.seismic.cfg")

function M.read()
	local f = io.open(path, "r")
	if not f then
		return nil
	end
	local content = f:read("*a")
	f:close()

	local api_key = content:match("api_key%s*=%s*(.-)\n") or content:match("api_key%s*=%s*(.-)$")
	local api_url = content:match("api_url%s*=%s*(.-)\n") or content:match("api_url%s*=%s*(.-)$")

	if not api_key or api_key == "" then
		return nil
	end
	return { api_key = vim.trim(api_key), api_url = api_url and vim.trim(api_url) or "" }
end

function M.write(cfg)
	local f = io.open(path, "w")
	if not f then
		vim.notify("Seismic: could not write " .. path, vim.log.levels.ERROR)
		return
	end
	f:write(string.format("[settings]\napi_key = %s\napi_url = %s\n", cfg.api_key, cfg.api_url))
	f:close()
end

return M