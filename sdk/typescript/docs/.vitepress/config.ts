import { defineConfig } from 'vitepress';

export default defineConfig({
  title: '@mirelth/synth',
  description: 'Synth TypeScript SDK Documentation',
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'API', link: '/api/' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Agent', link: '/guide/agent' },
            { text: 'Memory', link: '/guide/memory' },
            { text: 'Tools', link: '/guide/tools' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/Mirelth/Synth' },
    ],
  },
});
