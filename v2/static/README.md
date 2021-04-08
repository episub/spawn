Static is a library for managing access to external resources.  These resources may come from a few locations:

* A local folder
* (TODO) A remote internet URL

# Variants

Static can handle having a default folder as well as an override.  It will read
files from the variant folder, and then use the default folder if file isn't
found in the variant folder.

You may configure these two environment variables.

* STATIC_DEFAULT_FOLDER
* STATIC_VARIANT_FOLDER
