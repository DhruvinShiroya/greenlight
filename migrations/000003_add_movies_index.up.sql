/* add indexing for movie title and genres for better searching  */
CREATE INDEX IF NOT EXISTS movie_title_idx ON movies USING GIN (to_tsvector('simple',title));
CREATE INDEX IF NOT EXISTS movie_genres_idx ON movies USING GIN (genres);