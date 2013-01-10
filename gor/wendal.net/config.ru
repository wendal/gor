require 'rack'
require 'ruhoh'
run Ruhoh::Program.preview

# To preview your blog in "production" mode:
# run Ruhoh::Program.preview(:env => 'production')