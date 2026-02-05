.binary_name <- function() "plsdescribe"

.release_url <- function() {
  name <- .binary_name()
  base <- "https://github.com/b/plsdescribe/releases/download/nightly"
  os <- tolower(Sys.info()[["sysname"]])
  arch <- Sys.info()[["machine"]]

  suffix <- if (os == "darwin" && arch == "arm64") {
    "darwin-arm64"
  } else if (os == "darwin") {
    "darwin-amd64"
  } else {
    "linux-amd64"
  }

  paste0(base, "/", name, "-", suffix)
}

#' Find or download the plsdescribe binary
#'
#' Checks PATH first, then a per-user cache directory. Downloads from
#' GitHub releases if not found.
#'
#' @return Path to the binary.
#' @export
ensure_binary <- function() {
  name <- .binary_name()

  sys_path <- Sys.which(name)
  if (sys_path != "") return(unname(sys_path))

  cache_dir <- tools::R_user_dir("plsdescribe", which = "cache")
  if (!dir.exists(cache_dir)) dir.create(cache_dir, recursive = TRUE)

  dest <- file.path(cache_dir, name)

  if (!file.exists(dest)) {
    url <- .release_url()
    message("Downloading plsdescribe from ", url)
    utils::download.file(url, dest, mode = "wb", quiet = TRUE)
    Sys.chmod(dest, "0755")
    message("Installed to ", dest)
  }

  dest
}

#' Describe a plot image file
#'
#' @param image_path Path to an image file (PNG, JPEG, etc.)
#' @param verbose    Logical. Use detailed bullet-point description.
#' @param question   Optional follow-up question about the image.
#' @param tts        Logical. Speak the description via Google Cloud TTS
#'                   instead of returning text. Keeps stdout silent so it
#'                   won't collide with a screen reader.
#' @return Character vector of the description (invisibly when tts = TRUE).
#' @export
describe_image <- function(image_path, verbose = FALSE, question = NULL, tts = FALSE) {
  bin <- ensure_binary()

  args <- c("-f", image_path)
  if (verbose) args <- c(args, "-v")
  if (tts)     args <- c(args, "-tts")
  if (!is.null(question) && nzchar(question)) args <- c(args, "-q", question)

  tmp_out <- tempfile("plsdesc_", fileext = ".txt")
  on.exit(unlink(tmp_out), add = TRUE)
  args <- c(args, "-o", tmp_out)

  out <- system2(bin, args, stdout = TRUE, stderr = FALSE)

  if (tts) {
    return(invisible(out))
  }

  out
}

#' Describe a plot expression
#'
#' Renders a plot to a temporary PNG, then describes it.
#'
#' @param expr       A plot expression (base R) or a ggplot/lattice object.
#' @param verbose    Logical. Use detailed description.
#' @param question   Optional question about the plot.
#' @param tts        Logical. Speak instead of returning text.
#' @param width      PNG width in pixels.
#' @param height     PNG height in pixels.
#' @return Character vector of the description.
#' @export
describe_plot <- function(expr, verbose = FALSE, question = NULL, tts = FALSE,
                          width = 800, height = 600) {
  tmp_png <- tempfile("plot_", fileext = ".png")
  on.exit(unlink(tmp_png), add = TRUE)

  grDevices::png(filename = tmp_png, width = width, height = height)
  tryCatch({
    result <- force(expr)
    if (inherits(result, "ggplot") || inherits(result, "trellis")) {
      print(result)
    }
  }, finally = {
    grDevices::dev.off()
  })

  describe_image(tmp_png, verbose = verbose, question = question, tts = tts)
}
