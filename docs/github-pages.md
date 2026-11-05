# GitHub Pages product showcase

This repository includes a **static, public Property OS showcase** in `site/index.html`. It is a design-led product overview with a sample interactive portfolio tour, workflows, technical architecture and links to source/roadmap. Every example is invented and labelled illustrative.

**Important:** This is **not** a hosted Property OS application. GitHub Pages only serves static assets. Real users, session handling, PostgreSQL, S3 and the Go API require a separate backend deployment. Do not put API credentials, real tenant data, or internal URLs in the public showcase.

## Enable GitHub Pages

1. Open this GitHub repository's **Settings → Pages**.
2. Under **Build and deployment**, choose **GitHub Actions** as the source.
3. Merge the showcase pull request into `main`.
4. The **Property OS Pages showcase** workflow validates the HTML/JavaScript, uploads `site/`, and deploys using the protected `github-pages` environment.
5. Open the deployment URL in the workflow output and confirm the page works on desktop and mobile.

For the normal GitHub Pages repository path, the intended public URL is:

`https://realestatesassapplication.github.io/property-management-OS/`

**The URL is a deployment target, not proof it is online.** Verify the GitHub Actions deployment has succeeded before announcing it. Repository owners may need to approve Pages enablement or deploy permissions under the repository settings.

## Local preview

From the repository root:

```sh
python3 -m http.server 4173 --directory site
```

Open `http://localhost:4173`. No build or API keys are required.

Run static checks:

```sh
node --test scripts/check-pages-showcase.mjs
```

## Content safety and claims

- All visual dashboard figures and tab records are hardcoded fictitious examples; the page sends **no network API requests**.
- The product tour represents domain flows available in the codebase, not a production acceptance certificate.
- Owner Proofbook, reviewed CSV imports and Avara handoff are explicitly marked *planned*.
- Current implementation and limits are described in [feature status](feature-status.md), [architecture](architecture.md), and [production readiness](production-readiness.md).
- To show a **real interactive Property OS** deployment, use separate hosting for the Next.js server, Go API, PostgreSQL and S3; then add a clearly labelled and access-controlled application link to the showcase. Never embed demo/admin bearer tokens in public browser code.

## Workflow and rollout

The workflow is `.github/workflows/pages.yml`. Pull requests run a static check only. Pushes to `main` affecting the showcase deploy it. `workflow_dispatch` also supports manual redeployment.

An actual launch should include a final content review, external URL check, mobile/accessibility smoke tests, and a successful Pages deployment.
