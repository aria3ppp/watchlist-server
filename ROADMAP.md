# Watch-List Application

### Server side:

* User model
  * id
  <!-- * username
    -- index and unique -->
  * Email
  * Hashed password
  * First name
  * Last name
  * Birthdate
  * Joindate

<!-- * WatchLists -->
<!-- * Image Urls -->
<!-- * Scores (film, serie, artist, playlist, another user) -->

* [M2M] WatchList
  -- gui front-end should categorize films with shared season_id and serie_id together
  -- users can search for an artist and add one/all artist's films to their WatchList
  -- users can add one/all playlist films to their WatchList
  -- users can add one/all serie's or season's films to their WatchList
  * id
  * user_id  ---
                |--> index
  * film_id  ---
  * timestamp_added
  * timestamp_watched -- null if not watched

* [O2M] UserImageUrls
  * id
  * user_id
    -- index
  * image_url
  * timestamp

* [M2M] Scores_OutOf100
  * id
  * user_scoring_id
    -- non-null indexed
  * film_id
  * serie_id
  * artist_id
  * playlist_id
  * post_id
  * user_scored_id
  * timestamp

###################################

* [M2M] UserFollowing
  * id
  * user_following  ---
                       |--> index
  * user_followed   ---
  * timestamp

###################################

* [M2M] Posts
  -- Posts can contain text, media, link to other posts, users, artists, films, seasons, series, playlists, ...
  * id
  * user_id
    -- index
  * post_id
    -- non-null for post as comments
  * body
    -- format body contents as markdown?
  * timestamp_posted
  * timestamp_edited -- null if not edited

###################################

* Film (Movie/Episode) model
  * id
  * title
    -- index but not unique
  * descriptions
  * date_released
  * duration
  * episode_number
    -- null for movies
  * season_number
    -- null for movies
  * season_id
    -- null for movies
  * serie_id
    -- null for movies
  
  * contributed_by
  * contributed_at

  <!-- * Casts -->
  <!-- * MediaUrls -->

* Seasons
  -- Media (image, video) should fetched from serie's films
  -- Casts (actors, directors, screenwriters) should fetched from serie's films
  * id
  * title
    -- index
  * descriptions
  * season_number
  * date_started
  * date_ended
  * serie_id

  * contributed_by
  * contributed_at

* Serie model
  -- Media (image, video) should fetched from serie's films
  -- Casts (actors, directors, screenwriters) should fetched from serie's films
  * id
  * title
    -- index
  * descriptions
  * date_started
  * date_ended

  * contributed_by
  * contributed_at

* [M2M] FilmCasts 
  * id
  * film_id   ---
                 |--> index
  * artist_id ---
  * artist_role ('actor', 'director', 'screenwriter')

  * contributed_by
  * contributed_at

* [O2M] FilmMediaUrls
  * id
  * film_id
    -- index
  * media_url
  * media_type ('image', 'video')

  * contributed_by
  * contributed_at

* Artist (Cast/Director/Screenwriter) models
  * id
  * First name
  * Last name
  * Bio
  * birthdate

  * contributed_by
  * contributed_at

  <!-- * Images -->

* [O2M] ArtistImageUrls
  * id
  * artist_id
    -- index
  * image_url

  * contributed_by
  * contributed_at

* Playlist model
  -- cover image (like when creating a playlist on spotify it will auto-generate from movies/episodes)
  * Title
  * Descriptions
  * creator_user_id
  * timestamp

  <!-- * Films -->

* [M2M] PlaylistFilms
  * id
  * playlist_id ---
                   |--> index
  * film_id     ---

#############################################

* [] AdminUsers
  * id
  * user_id

#############################################

* AuditTrailContributionHistoryTable
  -- Record just update contribution history. Deletins are saved on main table and would tag IsDelted=TRUE
  -- For storage sake, record (e.g.) 50 last history of every table id
  -- Table ids should be indexed?

  * id
  * table_name
  * contributed_by
  * contributed_at

  * films_id
  * films_title
  * films_descriptions
  * films_date_released
  * films_episode_no
  * films_season_id

  * seasons_id
  * seasons_title
  * seasons_descriptions
  * seasons_season_no
  * seasons_serie_id
  * seasons_date_started
  * seasons_date_ended

  * series_id
  * series_title
  * series_descriptions
  * series_date_started
  * series_date_ended

  * film_casts_id
  * film_casts_film_id
  * film_casts_artist_id
  * film_casts_artist_role ('actor', 'director', 'screenwriter')

  * film_mediaurls_id
  * film_mediaurls_film_id
  * film_mediaurls_media_url
    -- don't remove image_url out of File Storage as long as AuditTable recordes this entry
  * film_mediaurls_media_type ('image', 'video')

  * artists_id
  * artists_first_name
  * artists_last_name
  * artists_bio
  * artists_birthdate

  * artist_imageurls_id
  * artist_imageurls_artist_id
  * artist_imageurls_image_url
    -- don't remove image_url out of File Storage as long as AuditTable recordes this entry