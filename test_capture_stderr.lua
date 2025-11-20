vim.cmd([[set runtimepath+=~/.config/nvim/pack/plugins/start/chuchu]])

local M = require('chuchu')
M.setup({})

vim.schedule(function()
  M.toggle_chat()
  vim.wait(200)
  
  local buf = vim.api.nvim_get_current_buf()
  local lc = vim.api.nvim_buf_line_count(buf)
  vim.api.nvim_buf_set_lines(buf, lc-1, lc, false, {'👤 | Analyse codebase and add pix payment'})
  
  local stderr_lines = {}
  local original_jobstart = vim.fn.jobstart
  vim.fn.jobstart = function(cmd, opts)
    local wrapped = vim.deepcopy(opts)
    wrapped.on_stderr = function(j, data, n)
      for _, line in ipairs(data or {}) do
        if line ~= '' then
          table.insert(stderr_lines, line)
          print('[STDERR] ' .. line)
        end
      end
      if opts.on_stderr then opts.on_stderr(j, data, n) end
    end
    return original_jobstart(cmd, wrapped)
  end
  
  M.send_message_from_buffer()
  
  vim.wait(5000)
  
  print('\n[STDERR SUMMARY]')
  for i, line in ipairs(stderr_lines) do
    print(i, line)
  end
  
  vim.cmd('qall!')
end)
