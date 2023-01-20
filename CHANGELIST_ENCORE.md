# CHANGE LIST
For run on non-Encore env, separate the Encore source from the original.

- [internal]
  - [config]
    - Edit [default.go] Change default web asset dir for use embed FS
    - Edit [helper.gen.go] Get default URL from Encore API
  - Edit [db/bundb/bundb.go] Connect to encore DB
  - [email]
    - Add [encore_util.go] Load template from embed FS
    - Edit [noopsender.go] Load from encore_util
  - [encore]
    - Add [migrations] Create Encore DB
    - Add [encore.go] For start server with Encore (from: /cmd/gotosocial/action/server/server.go)
  - [processing/account/create.go] Set first user as admin
  - [router]
    - Change `router` -> `RouterType`, `engine` -> `Engine` for use in encore/encore.go
    - Add [encore_router.go] `NewRouter` for use encore_template.go (from: router.go `New`)
    - Add [encore_template.go] `LoadTemplatesFromEmbed` Load from embed FS (reference: template.go)
  - [storage]
    - Add [terminusx] (reference: codeberg.org/gruf/go-store)
    - Edit [storage.go]
  - Edit [typeutils/defaulticons.go] Load icon from embed FS
  - Edit [web/assets.go] Use embed FS
- [web]
  - Generate [assets/dist] For embed FS
  - Edit [embed.go] For embed FS
- [go.mod]
  - Use Testify fork for skip test
