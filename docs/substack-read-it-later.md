# Read-it-later for Substack — feasibility investigation

**Question:** can a read-it-later flow in wui turn a Substack URL into a task with the
*right title* and a *duration*? Title matters most.

**Answer:** yes for the title — it is available in three independent places on every
Substack post, so a fallback ladder makes it near-certain. Duration is available too,
but it is not a field Substack publishes as "reading time": for text posts it must be
derived from the `wordcount` field, and for podcast/video posts it comes from
`podcast_duration`.

The real work is not the fetching. It is (a) that wui has no read-it-later feature
today, and (b) that wui's existing task-add path mangles titles containing `:` or `+`
— which describes a large share of Substack titles. See
[Blocker: `Add()` tokenisation](#blocker-add-tokenisation).

---

## Verification status

This investigation was done from a sandbox whose egress proxy blocks `substack.com`
(and most of the web), so **no claim below marked _unverified_ was tested against a
live Substack page.** Everything about the wui codebase *was* verified by reading the
code. A [probe script](#probe-script-run-this-first) is included so the unverified
claims can be confirmed in about a minute on a machine with normal internet access.

| Claim | Status |
|---|---|
| wui has no read-it-later feature | verified (grep across repo + full git history) |
| `Add()` splits the description on whitespace | verified (`internal/taskwarrior/client.go:137`) |
| `dur` UDA already parsed as a duration | verified (`internal/calendar/duration.go:28`) |
| yt-dlp's Substack extractor rejects text posts | verified (yt-dlp source, `yt_dlp/extractor/substack.py`) |
| `window._preloads = JSON.parse(...)` present in post HTML | verified (same yt-dlp source) |
| `GET /api/v1/posts/<slug>` returns a post object | documented by unofficial clients — **unverified** |
| That object contains `title`, `wordcount`, `podcast_duration` | documented — **unverified** |
| Cloudflare rejects a default Go user-agent | likely, by analogy with curl — **unverified** |

---

## Where the metadata lives

### 1. `GET https://<host>/api/v1/posts/<slug>` — best source

Substack serves an unauthenticated JSON endpoint per post. The slug is the last path
segment of a `/p/<slug>` URL. The unofficial Python client `NHagar/substack_api`
builds exactly `f"{base_url}/api/v1/posts/{slug}"` and sends only a browser
`User-Agent` header — no cookies, no key.

Relevant fields (as documented by clients built on the archive/post endpoints):

| Field | Use |
|---|---|
| `title` | the title, clean, no publication suffix |
| `subtitle` | optional, good annotation material |
| `wordcount` | the basis for reading time |
| `post_type` | `newsletter` / `podcast` / `video` |
| `podcast_duration` | seconds, for audio/video posts |
| `canonical_url` | de-duplication key |
| `audience` | `everyone` / `only_paid` — paywalled posts still expose title and wordcount |

### 2. `window._preloads` in the post HTML — same data, one extra hop

Every post page embeds `window._preloads = JSON.parse("<escaped JSON>")`, carrying the
same `post` object. yt-dlp relies on this and nothing else. Parsing it from Go needs
no HTML library: locate the `JSON.parse(` argument, `json.Unmarshal` it once as a
string to undo the JS escaping, then `json.Unmarshal` the result as the object.

### 3. `<meta property="og:title">` — the guaranteed floor

Always present, already clean (unlike `<title>`, which appends the publication name).
Needs `html.UnescapeString` for `&amp;` / `&#8217;`. Title only — no duration.

### 4. `https://<host>/feed` — for a poller, not for a saved URL

The RSS feed carries the ~20 most recent posts with full `content:encoded`, so title
*and* an exact word count are available without touching the API. This is the right
source if read-it-later ever grows a "subscribe to a publication" mode mirroring
`internal/youtube`'s playlist poller — but it cannot answer a question about an
arbitrary older URL.

### What about yt-dlp?

wui already shells out to yt-dlp (`internal/youtube/sync.go`), so reusing it is
tempting. It does not work here. Its Substack extractor:

- matches only `https?://[\w-]+\.substack\.com/p/<id>` — **custom domains do not match**
  (a large fraction of established Substacks use one, e.g. `newsletter.example.com`);
- raises `Page type "newsletter" is not supported` for any text post — i.e. exactly the
  read-it-later case;
- does not return `duration` even for the podcast posts it does accept.

Substack needs its own small fetcher.

## Duration

There is no "reading time" field. Two cases:

- **Text posts:** `reading_minutes = ceil(wordcount / wpm)`, `wpm` configurable,
  200 as the default, floor of 1 minute. Substack's own in-app estimate is computed
  the same way from `wordcount`.
- **Podcast / video posts:** `podcast_duration` (seconds) directly.

This lands nicely on an existing wui concept: `internal/calendar/sync.go` already reads
a **`dur` UDA** through `ParseTaskDuration`, which accepts ISO 8601 (`PT12M`) and
shorthand (`12min`). Writing `dur:PT12M` on a read-it-later task means a 12-minute
article automatically becomes a 12-minute block when it gets scheduled to Google
Calendar — no new concept, no new parser.

Accuracy caveat: `wordcount` for a paywalled post covers the full text, so a free
reader is shown a duration they cannot actually consume. Worth tagging
(`+paywalled`) rather than correcting.

---

## Blocker: `Add()` tokenisation

This is the part that actually needs a decision, and it is independent of Substack.

`internal/taskwarrior/client.go:136`:

```go
func (c *Client) Add(description string) (string, error) {
	descArgs := strings.Fields(description)
	args := append([]string{"add"}, descArgs...)
```

The description is split on whitespace and each token is handed to `task add` as a
separate argument, where Taskwarrior re-parses it. Any token that looks like
`name:value` or starts with `+`/`-` is consumed as an attribute or a tag rather than
description text. Substack titles routinely contain colons:

```
"Book Review: The Pale King"   ->  task add Book Review: The Pale King
"Highlights #47: what I read"  ->  #47: parsed as an attribute, + and - similar
```

So even a perfectly fetched title can be silently corrupted, or the add can fail
outright, on the most ordinary title shape there is. Every add path goes through this
one function — `internal/api/handlers.go:61`, `internal/gui/handlers.go:405`,
`internal/tui/model.go:2365`, `internal/youtube/sync.go:77`.

The fix is Taskwarrior's `--` separator, after which the rest of the command line is
treated as literal description text. That suggests a second method rather than a
change to `Add`'s behaviour, e.g.:

```go
// AddWithText creates a task whose description is taken literally, with attributes
// (project:, +tag, dur:) passed separately.
AddWithText(description string, attrs ...string) (string, error)
```

built as `task add <attrs...> -- <description>`. Existing callers keep the current
Taskwarrior-syntax-in-a-string behaviour; the read-it-later path uses the literal one.
**Confirm the `--` behaviour against the installed Taskwarrior version before
building on it** — it is documented, but it was not testable here.

---

## Proposed shape

A new `internal/readlater` package, sibling to `internal/youtube`, holding a
provider-based fetcher:

```go
type Metadata struct {
	Title    string
	Duration time.Duration // reading time, or podcast length
	Kind     string        // "newsletter" | "podcast" | "video"
	Paywalled bool
}

type Fetcher interface {
	Match(u *url.URL) bool
	Fetch(ctx context.Context, u *url.URL) (*Metadata, error)
}
```

with a Substack fetcher and a generic `og:title` fetcher as the terminal fallback.
Nothing new in `go.mod` — `net/http`, `encoding/json`, `regexp` and `html` cover it.

Substack fetcher ladder, each step falling through on failure:

1. `/api/v1/posts/<slug>` on the URL's **own host** (works for `*.substack.com` and
   custom domains alike, since Substack serves the API on both);
2. post HTML → `window._preloads`;
3. post HTML → `og:title` (title only, duration left unset).

Detection should not key on the `substack.com` hostname — that misses custom domains.
Cheapest robust test: attempt step 1 for any `/p/<slug>` URL and treat a non-JSON or
404 response as "not Substack", falling through to the generic fetcher.

Result becomes:

```
task add project:read +substack dur:PT12M -- Book Review: The Pale King
```

with the URL as an annotation (which keeps `o` — open-annotation-links — working) or
appended to the description, per existing convention.

### Practical requirements

- **User-Agent.** Substack is behind Cloudflare. Go's default `Go-http-client/1.1` is
  very likely rejected; every unofficial client sets a browser UA. Make it a config
  value so it can be changed without a rebuild.
- **Do not block the UI.** Add the task immediately with the URL, fetch in a
  goroutine, then `Modify` the description once metadata arrives. A 2–3s timeout and
  a graceful "URL only" outcome keeps a slow or blocked fetch from ever losing a save.
- **Test offline.** Save two or three real post pages and one API response as
  fixtures under `testdata/` and drive the parsers from them, the way
  `internal/taskwarrior/parser_test.go` does. That keeps the test suite free of
  network access — and would have made this investigation verifiable in-sandbox.

## Recommendation

Build the generic `og:title` fetcher first: it is a dozen lines, it works for every
site including Substack, and it delivers the thing that matters most. Add the Substack
API fast path second, purely to obtain `wordcount` for the duration. Fix the `Add()`
tokenisation before either, since neither is usable without it.

Risk is concentrated in Cloudflare's tolerance of a non-browser client — the one thing
here that Substack can change unilaterally, and the reason the ladder ends in a
fallback that only needs the HTML to arrive at all.

---

## Probe script (run this first)

Confirms the unverified claims in one go. Replace the URL with any real post.

```sh
URL="https://astralcodexten.substack.com/p/your-book-review-the-pale-king"
UA="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36"

HOST=$(printf '%s' "$URL" | cut -d/ -f1-3)
SLUG=${URL##*/}

# 1. does the JSON API answer unauthenticated, and does it carry the fields?
curl -sS -A "$UA" "$HOST/api/v1/posts/$SLUG" \
  | jq '{title, subtitle, post_type, wordcount, podcast_duration, audience, canonical_url}'

# 2. does a default (non-browser) user-agent get blocked?
curl -sS -o /dev/null -w 'default UA -> %{http_code}\n' "$HOST/api/v1/posts/$SLUG"

# 3. is window._preloads present in the HTML, and is og:title clean?
curl -sS -A "$UA" "$URL" | grep -o 'window\._preloads' | head -1
curl -sS -A "$UA" "$URL" | grep -o '<meta property="og:title"[^>]*>' | head -1

# 4. repeat 1 against a custom-domain Substack to confirm the API is served there too.
```

Step 2 answering `403` while step 1 answers `200` confirms the user-agent requirement.
Step 4 is the one that decides whether hostname-based detection is viable at all.
