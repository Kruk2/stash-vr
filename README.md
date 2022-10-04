# Stash-VR
Watch your [Stash](https://github.com/stashapp/stash) library in VR for that full immersion effect.

Stash-VR bridges your Stash instance and VR video player allowing you to browse, play and manage your scenes using the video players native VR UI.

It's light on resources, optionally configurable and has support for two-way sync.

**[Install](#Installation) Stash-VR, point it to your Stash instance and point your VR video player to Stash-VR.**

(Traditional flat/2d videos as well as VR videos are supported.)

## Supported video players
* HereSphere (two-way sync)
* DeoVR

## Features
* Browse, play and manage videos from your Stash library using native VR UI of supported video players.
* Display your Stash library as configured in Stash from your front page and saved filters.
* Provide transcoding endpoints to your videos served by Stash
* HereSphere
  * Two-way sync of studio, tags, performers, rating, favorites, markers
  * Increment o-count
  * Toggle organized flag
  * Generate categorized tags
  * Delete scenes
  * Funscript
* DeoVR
  * Markers

## Installation
Container images available at [docker hub](https://hub.docker.com/r/ofl0w/stash-vr/tags).

For details or docker compose see [docker_compose.yml](docker-compose.yml).

After installation open your endpoint (e.g. `http://localhost:9666`) in a regular browser to verify your setup.

### Docker
```
docker run --name=stash-vr -e STASH_GRAPHQL_URL=http://localhost:9999/graphql -e STASH_API_KEY=XXX -p 9666:9666 ofl0w/stash-vr:latest
```

Stash-VR listens on port `9666`, use docker port binding to change local port, e.g. `-p 9000:9666` to listen on port `9000` instead.

#### Configuration
* `STASH_GRAPHQL_URL`
  * **Required**
  * Url to your Stash graphql - something like `http://<stash.host>:<9999>/graphql`.
* `STASH_API_KEY`
  * Api key to your Stash if it's using authentication, otherwise not required.

<details>
<summary>Optional</summary>

* `FAVORITE_TAG`
  * Default: `FAVORITE`
  * Name of tag in Stash to hold scenes marked as [favorites](#favorites) (will be created if not present).
* `FILTERS`
  * Default: Empty
  * Narrow the selection of filters to show by setting one of below values:
    * `frontpage`
      * Show only filters found on Stash front page.
    * Comma separated list of filter ids, e.g. `1,5,12`
      * Show only filters from this list.
    * Empty
      * Show all saved filters.
* `ALLOW_SYNC_MARKERS`
  * Default: `false`
  * Enable sync of Marker from HereSphere [NOTE](#heresphere-sync-of-markers)
  * (Previously ~~`HERESPHERE_SYNC_MARKERS`~~, might be removed in the future.)
* `ENABLE_GLANCE_MARKERS`
  * Default: `false`
  * Put markers, instead of studio and tags, above the seekbar in HereSphere for quick access.
  * (Previously ~~`HERESPHERE_QUICK_MARKERS`~~, might be removed in the future.)
* `FORCE_HTTPS`
  * Default: `false`
  * Force Stash-VR to use HTTPS. Useful as a last resort attempt if you're having issues with Stash-VR behind a reverse proxy.
</details>

## Usage
Browse to `http://<host>:9666` using a supported video player. You'll be presented with your library within their respective native UI.
### HereSphere
##### Two-way sync
To enable two-way sync with Stash the relevant toggles (`Overwrite tags` etc.) in the cogwheel at the bottom right of preview view in HereSphere needs to be on.
#### Manage metadata
Scene metadata is handled using `Video Tags` in HereSphere.

To tag a scene open it in HereSphere and click `Video Tags` above the seekbar.
On any track insert a new tag and prefix it with `#:` i.e. `#:MusicVideo`.
This will create the tag `MusicVideo` in Stash if not already present and apply it to your scene. Removing a tag in HereSphere will untag the scene in Stash.

Same workflow goes for setting studio and performers but with different prefixes according to below:

|Metadata|Prefix| Alias        |
|--------|------|--------------|
|Tags|`#:`| `Tag:`       |
|Studio|`$:`| `Studio:`    |
|Performers|`@:`| `Performer:` |

#### Markers
(Both Stash and HereSphere use the word _tag_ but they use it differently. Tags in heresphere are akin to Markers in Stash)

Markers in Stash need a primary tag. Marker title is optional.
To create a marker using HereSphere play the target scene and create a "tag" on any track using `Video Tags`.
The naming format is:
* `<tag>:<title>` will create a Marker in Stash titled `<title>` with the primary tag `<tag>`
* `<tag>` will create a Marker in Stash with primary tag `<tag>` and no title.

Set the start time using HereSphere controls.
Tags (markers) in HereSphere has support for both a start and end time.
Stash currently defines Markers as having a start time only. This means the end time set in HereSphere will be ignored.

Enable sync of markers by setting `ALLOW_SYNC_MARKERS=true` but make sure you've also read the [caveat](#heresphere-sync-of-markers).

#### Favorites
When the favorite-feature of HereSphere is first used Stash-VR will create a tag in Stash named according to `FAVORITE_TAG` (set in docker env., defaults to `FAVORITE`) and apply that tag to your scene.

**Tip:** Create a filter using that tag, so it shows up in HereSphere for quick access to favorites.

#### Rating
HereSphere uses fractions for ratings, i.e. 4.5 is a valid rating. Stash uses whole numbers.
If you set a half star in HereSphere Stash-VR will round up the rating. That is if you set a rating of 3.5 the scene will receive a rating of 4 in Stash.
In other words, click anywhere on a star to set the rating to that amount of stars.

**Exception:** To remove a rating, rate the scene 0.5 (half a star).

#### O-counter
Increment o-count by adding a tag named `!O` (case-insensitive) in `Video Tags`.

Current o-count is shown as `O:<count>`

#### Organized
Toggle organized flag by adding a tag named `!Org` (case-insensitive) in `Video Tags`.

Current state is shown as `Org:<true/false>`

## VR
Both DeoVR and HereSphere has algorithms to automatically detect and handle VR videos.
It's not foolproof and to manually configure the players with custom layout/mesh-settings you can tag your scenes in Stash as follows:

* Mesh:
  - `DOME` 180° equirectangular
  - `SPHERE` 360° equirectangular
  - `FISHEYE` 180° fisheye
  - `MKX200` 200° fisheye
  - `RF52` 190° Canon fisheye
  - `CUBEMAP` Cubemap (lacks support in DeoVR?)
  - `EAC` Equi-Angular Cubemap (lacks support in DeoVR?)
* Layout:
  - `SBS` Side-by-side (Default)
  - `TB` Top-bottom

If a mesh is provided but no layout then default layout `SBS` will be used.

Most common combination is `DOME`+`SBS` meaning most VR videos only need the `DOME` tag.

## Known issues/Missing features
### Scene count limits
* DeoVR/HereSphere both seem to have limits and struggle/crash when too many videos are provided than they can handle.
  * For HereSphere the limit seems to be around 10k unique scenes.
  * Tip: If you have a VERY LARGE library and your player is struggling to load them all, try explicitly setting env. var. `FILTERS` with a list of filter ids such that the total amount of videos are lowered to a "reasonable" amount.

### Unsupported filter types
* Premade Filters (i.e. Recently Released Scenes etc.) from Stash front page are not supported. 
  * Tip: If you really want such filters to show they can easily be recreated and saved using regular filters in Stash.
### HereSphere sync of Markers
When using `Video Tags` in HereSphere to edit Markers Stash-VR will delete and (re)create them on updates.
There currently is no support for correlating the markers (tags) in HereSphere to a Marker in Stash.

**!! Any metadata, besides the primary tag and title, related to a marker will NOT be retained !!**\
(id, preview, secondary tags and created/updated time). If you're not using these fields anyway you probably won't notice the difference.

### Reflecting changes made in Stash
When the index page of Stash-VR is loaded Stash-VR will immediately respond with a cached version. At the same time Stash-VR will request the latest data and store it in the cache for the next request.
This means if changes are made in Stash and the player refreshed, it will receive the cached version built during the last (previous) request.
Just refresh again and the player should receive the latest changes. In other words, refresh twice.

### Stash version
Stash-VR has been tested against Stash v.0.16.1 - if you have issues arising from running an older version of Stash the recommended path is to upgrade Stash before attempting a fix.
