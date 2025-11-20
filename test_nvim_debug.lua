vim.cmd([[set runtimepath+=~/.config/nvim/pack/plugins/start/chuchu]])

local M = require('chuchu')
M.setup({})

print("[TEST] Starting simulation")

vim.schedule(function()
  M.toggle_chat()
  vim.wait(200)
  
  print("[TEST] Sending message")
  local buf = vim.api.nvim_get_current_buf()
  local lc = vim.api.nvim_buf_line_count(buf)
  vim.api.nvim_buf_set_lines(buf, lc-1, lc, false, {'👤 | Analyse codebase and add pix payment'})
  
  M.send_message_from_buffer()
  
  vim.wait(8000)
  
  print("[TEST] Checking events file")
  local events = vim.fn.readfile(vim.fn.expand("~/.chuchu/events.jsonl"))
  print("[TEST] Events in file:", #events)
  for i, e in ipairs(events) do
    print("  " .. i .. ": " .. e:sub(1, 60))
  end
  
  vim.cmd('qall!')
end)
