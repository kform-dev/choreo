/*
Copyright 2024 Nokia.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

type IPFamilyPolicy string

const (
	// IPFamilyPolicyNone defines no L3 addressing, meaning L2
	IPFamilyPolicyNone IPFamilyPolicy = "none"
	// IPFamilyPolicyIPv4Only defines L3 IpFamilyPolicy as ipv4 only
	IPFamilyPolicyIPv4Only IPFamilyPolicy = "ipv4only"
	// IPFamilyPolicyIPv6Only defines L3 IPFamilyPolicy as ipv6 only
	IPFamilyPolicyIPv6Only IPFamilyPolicy = "ipv6only"
	// IpFamilyPolicyDualStack defines L3 IpFamilyPolicy as dual stack (ipv4 and ipv6)
	IPFamilyPolicyDualStack IPFamilyPolicy = "dualstack"
)
