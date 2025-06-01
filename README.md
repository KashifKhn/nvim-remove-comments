# nvim-remove-comments

A simple Neovim plugin to **remove all comments** from the current buffer using Tree-sitter.

> Great for stripping LLM-generated comments or cleaning up code.

![Showcase](example.gif)

---

## ✨ Features

- Removes all comment nodes from the current buffer
- Tree-sitter powered: works reliably across languages
- Fast and minimal
- Works with many filetypes (JS, TS, Lua, HTML, Python, etc.)

---

## 🧠 Why?

Sometimes you want to clean up your file by removing all comments — especially if you're dealing with auto-generated or outdated ones. This plugin gives you one keybinding to do just that.

---

## ⚡ Installation

### [lazy.nvim](https://github.com/folke/lazy.nvim)

```lua
{
  "KashifKhn/nvim-remove-comments",
  config = function()
    require("nvim-remove-comments").setup()
  end,
}
```
