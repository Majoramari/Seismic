local M = {}

local global_config = require("seismic.global_config")

M.options = {
	api_url = "https://correct-wolverine-majoramari-6049fd71.koyeb.app",
	enabled = true,
	statusline_enabled = true,
}

function M.setup(opts)
	M.options = vim.tbl_deep_extend("force", M.options, opts or {})
end

-- API key is never stored in options table on purpose — always read fresh
-- from the shared ~/.seismic.cfg so key rotation in another editor is
-- picked up immediately without needing :SeismicReload.
function M.get_api_key()
	local cfg = global_config.read()
	return cfg and cfg.api_key or ""
end

function M.get_api_url()
	local cfg = global_config.read()
	if cfg and cfg.api_url and cfg.api_url ~= "" then
		return cfg.api_url
	end
	return M.options.api_url
end

function M.has_api_key()
	return M.get_api_key() ~= ""
end

function M.is_enabled()
	return M.options.enabled
end

function M.set_enabled(value)
	M.options.enabled = value
end

function M.set_api_key(key)
	global_config.write({ api_key = key, api_url = M.get_api_url() })
end

return M