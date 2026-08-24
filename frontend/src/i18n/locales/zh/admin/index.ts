import overview from './overview'
import channels from './channels'
import accounts from './accounts'
import resources from './resources'
import ops from './ops'
import settings from './settings'
import audit from './audit'
import promptAudit from './promptAudit'
import authIPBan from './authIPBan'
import checkin from './checkin'
import emailBroadcasts from './emailBroadcasts'
import plugins from './plugins'

export default {
  ...overview,
  ...channels,
  ...accounts,
  ...resources,
  ...ops,
  ...settings,
  ...audit,
  ...promptAudit,
  ...authIPBan,
  ...checkin,
  ...emailBroadcasts,
  ...plugins,
}
