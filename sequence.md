# What do we need? 

1. Serve images in WebP format
2. Track number of hits per image
3. Log the hits in the database

## What's acutally happening in `main`

### `main()`

1. create a seed for the random functions - probably not optimizable
2. log the db name - unnecessary but we'll leave it for now
3. open the db - no opt
4. create the table - maybe opt? 
5. assign `handleRoot()` to "/" endpoint - no opt yet
6. log port listening message - no opt
7. `http.ListenAndServe` call - no opt

### `handleRoot()`

Whenever a GET request is made to "/", `handleRoot()` is called:

1. `handleRoot()` returns an anonymous function. It does the following: 
2. calls `getImage()` and returns a name and img
3. calls `IncrementHitCount()` to increment the hit count on the image
4. logs "Serving [img name]" - can be eliminated
5. writes the http response as a string of bytes - needs furter research

The two program functions `getImage()` and `IncrementHitCount()` need further exploration:

#### `IncrementHitCount()`

1. updates the database - perhaps query language could be optimized???
2. displays an error message if appropriate (and calls a package to turn the error message red) - optimization might include removing the color change mechanism but may not be worthy

#### `getImage()`

`getImage()` returns a string and a WebpImage

1. instantiates string variable `imgBase` with value "data"
2. instantiates `imgName` as the result of calling `getImageNames()` and passes `imgBase` - passing string directly would eliminate 1; additionally we should only have to call `getImageNames()` once and cache the results;
3. instantiates `index` with a random number in the range of the length of the slice returned by `getImageNames()` which itself is a constant and could be cached
4. instantiates `imgEntry` which is a filename for a file in `imgName` slice
5. instantiates `imgPath` which creates a path for the image
6. instantiates `rawImg` calls `loadImg()` to parse the image
7. instantiates `webpImg` which converts the raw image to webp format
8. returns the img name (a string) and the image itself (webp) to `handleRoot()`

Questions: 
    1. can we convert directly from png to webp without parsing and encoding
    2. can we optimize the randomization process
    3. can the gathering of the filenames and selecting an image be done outside of the the handler? 

## Suggested optimizations in overall sequence

1. Move most of the functionality of `getImage()` outside the handler
2. Cache a large slice of filenames on startup - goroutine
3. separate the logging of the hits from the handler - it can go in a separate goroutine and is not intrinsically tied to the serving of the images
4. streamline the functionality in `getImage()` by eliminating unnecessary variables and removing repeated operations
5. streamline the image conversion process if possible - can be pre-converted, no need to wait until they're served
6. the handler itself only needs to return the the next image in the queue and kick off an incrementing goroutine

#### Image conversion process

1. the file is read as a string of bytes by `loadImg()`
2. it's converted to an image.Image (using image.Decode()`)by `parseImg()`
3. it's converted (using `webp.Enconde()`) by `imgToWebp()` 

This process may not be optimized further...?

## Suggested New Sequence

### `main()`

1. create seed
2. connect to DB
3. goroutine1 - create table
4. goroutine2 - convert all images to webp
5. goroutine2 - get paths from "data"
6. goroutine2 - create slice of filenames of length 6000
7. handle "/"
8. listen and serve

### `handleRoot()`

1. check i
2. goroutine1 - serve image[i]
3. goroutine2 - increment hit
4. log (remove after testing)

### `getImage()`

1. (functionality moved to `main()` goroutine2)
2. gets an image based on randomNSlice[i] and imagePathsList[j]
3. returns name and path
4. increments i