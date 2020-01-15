/** @format */

export const Empty = ({title, subtitle, children}) => (
  <div class="empty">
    <div class="empty-icon">
      <i class="icon icon-apps"></i>
    </div>
    <p class="empty-title h5">{title}</p>
    <p class="empty-subtitle">{subtitle}</p>
    <div class="empty-action">{children}</div>
  </div>
)
