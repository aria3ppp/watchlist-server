SELECT  id,
        title,
        descriptions,
        date_released,
        duration,
        poster,
        contributed_by,
        contributed_at,
        invalidation
FROM films
WHERE 
        contributed_at > ?
    AND
        series_id IS NULL
    AND
        season_number IS NULL
    AND
        episode_number IS NULL
ORDER BY
    contributed_at ASC;