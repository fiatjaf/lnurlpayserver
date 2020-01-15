/** @format */

export const TileCompact = ({
  title = 'Title',
  subtitle = 'Sub',
  icon = 'icon-resize-horiz',
  image = null
}) => (
  <div class="column col-12">
    <div class="tile tile-centered">
      <div class="tile-icon">
        <div class="example-tile-icon">
          {image ? (
            <img src={image} alt="image" />
          ) : (
            <i class={`icon ${icon} centered`}></i>
          )}
        </div>
      </div>
      <div class="tile-content">
        <div class="tile-title">{title}</div>
        <p class="tile-subtitle text-gray">{subtitle}</p>
      </div>
    </div>
  </div>
)
