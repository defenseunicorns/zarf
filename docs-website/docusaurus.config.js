// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const prismRenderer = require("prism-react-renderer/dist/index");
const { SocialsBox } = require("./static-components/SocialsBox/SocialsBox");

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Zarf Documentation",
  tagline: "Airgap is hard. Zarf makes it easy.",
  url: "https://zarf.dev",
  baseUrl: "/",
  onBrokenLinks: "warn",
  onBrokenMarkdownLinks: "throw",
  favicon: "img/favicon.svg",
  organizationName: "Defense Unicorns", // Usually your GitHub org/user name.
  projectName: "Zarf", // Usually your repo name.
  markdown: {
    mermaid: true,
  },
  themes: [
    [require.resolve("@easyops-cn/docusaurus-search-local"), { hashed: true }],
    [require.resolve("@docusaurus/theme-mermaid"), { hashed: true }],
  ],
  staticDirectories: ["static", "../examples", "../packages"],
  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          path: "..",
          include: [
            "CONTRIBUTING.md",
            "adr/**/*.{md,mdx}",
            "docs/**/*.{md,mdx}",
            "examples/**/*.{md,mdx}",
            "packages/**/*.{md,mdx}",
          ],
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
          sidebarPath: require.resolve("./src/sidebars.js"),
          editUrl: ({ docPath }) => {
            // TODO: (@RAZZLE) once examples have been fixed, change this url to edit: `https://github.com/defenseunicorns/zarf/edit/main/${docPath}`
            return `https://github.com/defenseunicorns/zarf/tree/main/${docPath}`; // <-- to view
          },
          routeBasePath: "/",
          async sidebarItemsGenerator({ defaultSidebarItemsGenerator, ...args }) {
            const sidebarItems = await defaultSidebarItemsGenerator(args);
            if (args.item.dirName === "docs") {
              // This hack places the examples tree at the 7th position in the sidebar
              sidebarItems.splice(6, 0, {
                type: 'category',
                label: 'Package Examples',
                link: {
                  type: "doc",
                  id: "examples/README",
                },
                items: [
                  {
                    type: "autogenerated",
                    dirName: "examples",
                  },
                ],
              });
            }
            if (args.item.dirName === "examples") {
              // This hack removes the "Overview" page from the sidebar on the examples page
              return sidebarItems.slice(1);
            }
            return sidebarItems;
          },
        },
        blog: false,
        theme: {
          customCss: [require.resolve("./src/css/custom.css")],
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: {
        defaultMode: "dark",
        disableSwitch: true,
      },
      navbar: {
        logo: {
          alt: "Zarf",
          src: "img/zarf-logo-light.svg",
          srcDark: "img/zarf-logo-dark.svg",
          href: "https://zarf.dev/",
          target: "_self",
        },
        items: [
          {
            type: "search",
            position: "right",
          },
          {
            type: "doc",
            docId: "docs/zarf-overview",
            position: "left",
            label: "Docs",
          },
          {
            position: "left",
            label: "Product",
            to: "https://zarf.dev",
            target: "_self",
          },
          {
            type: "html",
            position: "right",
            className: "navbar__item--socials-box",
            value: SocialsBox({
              linkClass: "menu__link",
            }),
          },
        ],
      },
      footer: {
        style: "dark",
        logo: {
          alt: "Zarf",
          src: "img/zarf-logo-light.svg",
          srcDark: "img/zarf-logo-dark.svg",
          href: "https://zarf.dev/",
        },
        copyright: `<p class="p-copy">Copyright © ${new Date().getFullYear()} Zarf Project, All rights reserved.</p>`,
        links: [
          {
            html: SocialsBox(),
          },
        ],
      },
      prism: {
        theme: prismRenderer.themes.shadesOfPurple,
        darkTheme: prismRenderer.themes.shadesOfPurple,
      },
    }),
};

module.exports = config;
