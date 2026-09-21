# Embedded Image Studio

Source: https://github.com/aboutnb/gpt_image_playground
Pinned commit: da4fda85b59ecacc51d6a1e2ef680e3abb9e29b8
License: MIT, retained in LICENSE.

This sub-application is built independently with npm. StudioApp replaces the
standalone entry, and studioBridge sends only allowlisted image requests to the
authenticated parent. No credentials are configured or persisted here.

Build after the Vue frontend: npm ci && npm run build. The build output goes to
../backend/internal/web/dist/image-studio-app/. The root Docker build includes it.
The app must be opened through /image-studio, not as a standalone application.
