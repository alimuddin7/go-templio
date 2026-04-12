/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./views/**/*.templ",
    "./internal/transport/http/handler/**/*.go",
    "./node_modules/**/*.templ", // if vendored
    // We add the actual package path from GOMOD as well if needed, 
    // but usually standard ./views is enough for custom components.
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: '#7c6aff',
          dim: '#5e50d6',
        },
        surface: {
          0: '#0d0f14',
          1: '#13161e',
          2: '#1a1e2a',
          3: '#21263a',
          4: '#2a3048',
        }
      },
      borderRadius: {
        'md': '10px',
        'lg': '16px',
        'xl': '24px',
      }
    },
  },
  plugins: [],
}
