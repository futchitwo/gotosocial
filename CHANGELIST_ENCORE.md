# CHANGE LIST
For run on non-Encore env, separate the Encore source from the original.

- [internal]
  - [config]
    - [default.go] Change default web asset dir for use embed FS
    - [helper.gen.go] Get default URL from Encore API
  - [db/bundb/bundb.go] Connect to encore DB
  - [email]
    - [encore_util.go] Load template from embed FS
    - [noopsender.go] Load from encore_util
  - [encore]
    - [migrations] Create Encore DB (from: /cmd/gotosocial/action/server/server.go)
    - [encore.go] For start server with Encore
  - [processing/account/create.go] Set first user as admin
  - [router]
    - Change `router` -> `RouterType`, `engine` -> `Engine` for use in encore/encore.go
    - [encore_router.go] `NewRouter` for use encore_template.go (from: router.go `New`)
    - [encore_template.go] `LoadTemplatesFromEmbed` Load from embed FS (reference: template.go)
  - [typeutils/defaulticons.go] Load icon from embed FS
  - [web/assets.go] Use embed FS
- [web]
  - [assets/dist] For embed FS
  - [embed.go] For embed FS
- [go.mod]
  - Use Testify fork for skip test
