local M = {}

function M.setup(opts)
	require("seismic.config").setup(opts)
	require("seismic.commands").setup()

	local heartbeat = require("seismic.heartbeat")
	local augroup = vim.api.nvim_create_augroup("Seismic", { clear = true })

	vim.api.nvim_create_autocmd({ "TextChanged", "TextChangedI" }, {
		group = augroup,
		callback = function(args)
			heartbeat.handle_activity(args.buf, false)
		end,
	})

	vim.api.nvim_create_autocmd("BufEnter", {
		group = augroup,
		callback = function(args)
			heartbeat.handle_activity(args.buf, true)
		end,
	})

	vim.api.nvim_create_autocmd("BufWritePost", {
		group = augroup,
		callback = function(args)
			heartbeat.handle_activity(args.buf, true)
		end,
	})

	vim.api.nvim_create_autocmd("FocusGained", {
		group = augroup,
		callback = function()
			heartbeat.handle_activity(vim.api.nvim_get_current_buf(), false)
		end,
	})

	-- Retry queued heartbeats every 5 minutes
	local timer = vim.uv.new_timer()
	timer:start(300000, 300000, vim.schedule_wrap(function()
		heartbeat.flush_queue()
	end))
end

return M