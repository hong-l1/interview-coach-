import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      colors: {
        ink: '#172033',
        paper: '#f7f4ed',
        moss: '#4f6f52',
        coral: '#d96c4a',
        cobalt: '#3157b7',
      },
      boxShadow: {
        soft: '0 18px 60px rgba(23, 32, 51, 0.12)',
      },
    },
  },
  plugins: [],
};

export default config;
