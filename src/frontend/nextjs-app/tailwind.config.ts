import type { Config } from 'tailwindcss'

// Design tokens are the default Tailwind scale — docs/06-frontend/design-system.md:
// "Token system xây dựng trực tiếp trên Tailwind default scale, không cần
// custom tailwind.config" — so this file only wires up content globs.
const config: Config = {
  content: ['./app/**/*.{ts,tsx}', './components/**/*.{ts,tsx}', './lib/**/*.{ts,tsx}'],
  theme: {
    extend: {},
  },
  plugins: [],
}

export default config
