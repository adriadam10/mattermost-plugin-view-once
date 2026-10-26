# Mattermost View Once plugin

WhatsApp-style view-once messages for Mattermost: text and any files that each recipient can open only once.

- A button in the message box toolbar opens the composer. It takes the text already in the box; files are added with the picker, by dropping them or by pasting them (any type, up to Mattermost's *Maximum file size*, at most 10).
- The channel only gets a "View once" card. Each member can open it **once**; when they close the viewer it cannot be opened again. A viewer left open closes on its own after 30 minutes.
- The message is deleted for good when every member (bots and deactivated users aside) has opened and closed it, or after *Maximum lifetime* (System Console → Plugins → View Once, default 7 days, 1 to 30).
- As in WhatsApp, the sender cannot open it again once sent. They can see who opened it and when, and delete it for everyone. Deleting the post from its `···` menu deletes the content too.
- Works in direct messages, group messages, channels and threads.

## How it works

The content never touches Mattermost posts or file attachments, so it does not
show up in notifications, search, exports, link previews or the channel files:

1. The composer creates a draft with the text (`POST /api/v1/messages`) and uploads
   each file as a raw body (`POST /api/v1/messages/{id}/files`). Files are stored
   in the plugin KV store in 4 MB chunks.
2. Sending it (`POST /api/v1/messages/{id}/send`) creates a post of type
   `custom_view_once` whose text is only a notice, with the message id in its props.
   Posts of that type created any other way are rejected.
3. Opening the card (`POST /api/v1/messages/{id}/open`) records the member's only
   view and returns the content. The sender is refused. Files can be downloaded while the viewer is open.
4. Closing it (`POST /api/v1/messages/{id}/close`) ends the view. If everybody is done,
   the chunks, the message and the post are deleted.

A background job every 5 minutes deletes expired messages, drafts never sent (after
an hour) and messages whose last viewer never closed the viewer. Every KV entry also
expires on its own a day after its message, in case the job misses it.

Deleting removes the content from the database; database backups taken while the
message existed still hold it until they are rotated.

## Limitations

- The native mobile apps do not run webapp plugins: they show the notice, which
  asks to open the message from the web or desktop app.
- A view-once message cannot be edited (the post keeps its notice) or forwarded.

## Build

Only Docker is needed:

```sh
./build.sh
# -> dist/com.github.adriadam10.view-once-<version>.tar.gz
```

## CI and releases

- Every push and PR runs the Go and webapp tests and the build; the plugin `.tar.gz` is attached to the run as an artifact.
- Pushing a tag `vX.Y.Z` (matching `version` in `plugin.json`) publishes a GitHub release with the plugin.
- Dependabot checks npm, Go modules and GitHub Actions weekly. Patch/minor updates merge automatically once CI passes. The Mattermost plugin API module is pinned on purpose: it must not get ahead of the oldest supported server.

## Install

```sh
mmctl plugin install-url https://github.com/adriadam10/mattermost-plugin-view-once/releases/download/vX.Y.Z/com.github.adriadam10.view-once-X.Y.Z.tar.gz
mmctl plugin enable com.github.adriadam10.view-once
```

Uploading requires `PluginSettings.EnableUploads = true`.

Tested on Mattermost 12.0.

## License

MIT
