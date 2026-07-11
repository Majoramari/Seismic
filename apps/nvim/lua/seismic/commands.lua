local M = {}

local config = require("seismic.config")
local heartbeat = require("seismic.heartbeat")

function M.setup()
	vim.api.nvim_create_user_command("SeismicSetApiKey", function()
		vim.ui.input({ prompt = "Seismic API key: " }, function(input)
			if input and input ~= "" then
				config.set_api_key(input)
				vim.notify("Seismic: API key saved!", vim.log.levels.INFO)
			end
		end)
	end, {})

	vim.api.nvim_create_user_command("SeismicEnable", function()
		config.set_enabled(true)
		vim.notify("Seismic: Tracking enabled", vim.log.levels.INFO)
	end, {})

	vim.api.nvim_create_user_command("SeismicDisable", function()
		config.set_enabled(false)
		vim.notify("Seismic: Tracking disabled", vim.log.levels.INFO)
	end, {})

	vim.api.nvim_create_user_command("SeismicOpenDashboard", function()
		vim.ui.open("https://seismic.icu/dashboard")
	end, {})

	vim.api.nvim_create_user_command("SeismicStatus", function()
		if not config.has_api_key() then
			print("Seismic: no API key set. Run :SeismicSetApiKey")
		elseif not config.is_enabled() then
			print("Seismic: tracking paused")
		else
			print("Seismic: tracking active")
		end
	end, {})
end

return M