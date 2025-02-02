package preply

import (
	"fmt"
)

func Fetch() error {
	cookie := "has_account=true; uid=3412a446bc7ca92809401ab127901e0085ba02291dcda3e7b2bdb93577e7b8ba; exp_usercentrics_web_stickybanner=1; _gcl_au=1.1.1662069986.1735660585; _ga=GA1.1.1283477354.1735660586; _fbp=fb.1.1735660586112.34775148036686386; hubspotutk=80267cee5d29806af78e2f6742be2a19; _tt_enable_cookie=1; _ttp=0G5pHGNVKtZw1nPqwZRaA0_Mnnr.tt.1; intercom-device-id-d97m90f7=7024d273-13b8-40d7-9dc9-08c6c975f450; __stripe_mid=f8d3ade7-e193-4be2-be82-9e3d863492a55d8cfa; _hjSessionUser_641144=eyJpZCI6Ijk5ZmU2MDRlLWVkNjctNThhYS1iOGFkLTVkNDdkZDk2ZGZlOCIsImNyZWF0ZWQiOjE3MzU2NjA1ODUxMzQsImV4aXN0aW5nIjp0cnVlfQ==; _hjMinimizedPolls=1558378; _hjDonePolls=1558378%2C1561820%2C1558665; hj_first_visit_30days=2025-01-23T15:57:18.735Z; currency_code=PLN; __cf_bm=E1J4DNKWIX6J.660GtBf_HoFcmp4LeIOrwnwPiGe9nQ-1737802607-1.0.1.1-9nsJZAxR4H5hqPNhr2EDa5eeAWcGa_z8bYFYmpxyTLDqK7k7COniE6fAH6duveiIFG2C0B72gInU8kaiEdCgMA; _cfuvid=UOazmfLeAPXQzt6bSLrxMexl0w1SV1pbRVo8q4FJhJs-1737802607332-0.0.1.1-604800000; m_source=other; m_source_landing=/; m_source_details=; is_source_set=yes; source_page=; landing_page=https://preply.com/; visit_time=2025-01-25T10:56:47.689Z; browserTimezone=Europe/Warsaw; __hstc=115815577.80267cee5d29806af78e2f6742be2a19.1735660586125.1737651250637.1737802609128.13; __hssrc=1; _hjSession_641144=eyJpZCI6IjFhYjYzM2NmLTE1ZGYtNDE0NC1iOWM4LWQyM2U5OGQ4MGViNiIsImMiOjE3Mzc4MDI2MTA2NDUsInMiOjAsInIiOjAsInNiIjowLCJzciI6MCwic2UiOjAsImZzIjowLCJzcCI6MH0=; init_uid=3412a446bc7ca92809401ab127901e0085ba02291dcda3e7b2bdb93577e7b8ba; banner_support_ua_seen=1737802617425; user_id=11816321; last_landing_url=\"https://preply.com/complete/google-oauth2/?state=EWR0V1xIEdJ0NQG9RbNqLRoMETgDXWhO&code=4%2F0ASVgi3KxogdsYAe203f3-SRiKViGbHEfduciKNOX7Enfx1C3SBrfxkmgLYpDiv-G7g-N2w&scope=email+profile+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.profile+openid+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.email&authuser=0&prompt=none\"; last_landing_referrer=\"https://accounts.google.com/\"; x-cf-auth=true; sessionid=c3vml1eehp5nnh48rb3zrbi5hd8fk6lx; ssr_block_cached_logo=true; __stripe_sid=9f35b339-1541-40ca-ad4b-3f34b92885206f5618; ab.storage.deviceId.ef180e24-d095-41d8-8eb6-1b498c6dcdc0=%7B%22g%22%3A%226fee69b7-cdf9-98c4-cd93-28dfae89875d%22%2C%22c%22%3A1722244373924%2C%22l%22%3A1737802990275%7D; ab.storage.userId.ef180e24-d095-41d8-8eb6-1b498c6dcdc0=%7B%22g%22%3A%2211816321%22%2C%22c%22%3A1722244373919%2C%22l%22%3A1737802990276%7D; _dd_s=logs=1&id=f8cba2dc-e6bc-46de-a9d1-d9fbce2d4d6f&created=1737802700907&expire=1737803897887; _dd_l=1; _dd=97ea015b-a17c-4773-97df-05fe71acf26a; pv_count=11; _uetsid=13832f10db0b11efa66a5d1ff8249124; _uetvid=5f060bf04d8a11ef844e198e670485b7; __hssc=115815577.11.1737802609128; ab.storage.sessionId.ef180e24-d095-41d8-8eb6-1b498c6dcdc0=%7B%22g%22%3A%22a3a79660-3d45-d34a-0708-e733f6c15399%22%2C%22e%22%3A1737803305211%2C%22c%22%3A1737802990274%2C%22l%22%3A1737803005211%7D; _hjHasCachedUserAttributes=true; _ga_BQH4D3BLSB=GS1.1.1737802608.15.1.1737803037.3.0.0; csrftoken=JzKWkflGwii57ukbGAhLk1AIf3DuwnSSbptuaeYsVS22DUsYthVmUYfxU8C4sfrj"

	f := VocabFetcher{}
	v, err := f.Fetch(cookie, 1, 0)
	if err != nil {
		return err
	}

	fmt.Printf("VocabRes: %#v", v)
	return nil
}
