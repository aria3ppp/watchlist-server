SELECT  id,
        title,
        descriptions,
        date_started,
        date_ended,
        contributed_by,
        contributed_at,
        invalidation
FROM serieses
WHERE 
    contributed_at > ?
ORDER BY
    contributed_at ASC;