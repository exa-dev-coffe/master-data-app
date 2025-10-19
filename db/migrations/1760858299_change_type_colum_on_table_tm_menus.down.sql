ALTER TABLE public.tm_menus
    ALTER COLUMN rating TYPE double precision USING rating::double precision;