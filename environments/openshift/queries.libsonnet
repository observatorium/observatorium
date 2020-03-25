{
  queries: [
    {
      name: 'Clusters',
      query: 'avg_over_time(sum(count by (_id) (max without (prometheus,receive,instance) ( cluster_version{type="current"} )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h])',
    },
    {
      name: 'Clusters aged 1w',
      query: 'avg_over_time(sum(count by (_id) (max without (prometheus,receive,instance) ( (time() - cluster_version{type="initial"}) > (7 * 24 * 60 * 60) )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h])',
    },
    {
      name: 'Nodes',
      query: 'avg_over_time(sum(sum by (_id) (max without (prometheus,receive,instance) ( cluster:node_instance_type_count:sum)) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h])',
    },
    {
      name: 'Cores',
      query: 'avg_over_time(sum(sum by (_id) (max without (prometheus,receive,instance) ( cluster:capacity_cpu_cores:sum)) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h])',
    },
    {
      name: 'Workload CPU',
      query: 'avg_over_time(sum(max by (_id) (max without (prometheus,receive,instance) ( workload:cpu_usage_cores:sum )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h])',
    },
    {
      name: 'Etcd Objects',
      query: 'avg_over_time(sum(sum by (_id) (max without (prometheus,receive,instance) ( instance:etcd_object_counts:sum )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h])',
    },
    {
      name: 'Weekly Active Users',
      query: 'count(count by (account) (count_over_time(subscription_labels{email_domain!~"redhat.com|(.*\\\\.|^)ibm.com"}[7d])))',
    },
    {
      name: 'Unique customers',
      query: 'count(count by (email_domain) (count_over_time(subscription_labels{email_domain!~"redhat.com|(.*\\\\.|^)ibm.com|gmail.com"}[7d])))',
    },
    {
      name: 'Hybrid customers',
      query: 'count(count by (email_domain) (count by (email_domain,type) (count by (_id,type,email_domain) (cluster_infrastructure_provider{} + on (_id) group_left(email_domain) (topk by (_id) (1, 0 * subscription_labels{}))))) and on (email_domain) (count by (email_domain) (count by (email_domain,type) (count by (_id,type,email_domain) (cluster_infrastructure_provider{} + on (_id) group_left(email_domain) (topk by (_id) (1, 0 * subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com|gmail.com"}))))) > 1))',
    },
    {
      name: 'Subscribed clusters',
      query: 'avg_over_time(sum(count by (_id) (max without (prometheus,receive,instance) ( cluster_version{type="current"} )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{support=~"Standard|Premium|Layered"})))[7d:12h])',
    },
    {
      name: 'Subscribed nodes',
      query: 'avg_over_time(sum(sum by (_id) (max without (prometheus,receive,instance) ( cluster:node_instance_type_count:sum)) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{support=~"Standard|Premium"})))[7d:12h])',
    },
    {
      name: 'Subscribed cores',
      query: 'avg_over_time(sum(sum by (_id) (max without (prometheus,receive,instance) ( cluster:capacity_cpu_cores:sum)) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{support=~"Standard|Premium|Layered"})))[7d:12h])',
    },
    {
      name: 'Hours failing per week',
      query: 'sum((max by (_id) (count_over_time((cluster_version{type="failure"} * 0 + 1)[7d:15m]) > 1) + on (_id) group_left(email_domain) topk by (_id) (1, 0 * max_over_time(subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"}[7d]))) / 4)',
    },
    {
      name: 'Average code age (days)',
      query: 'avg_over_time(avg(max by (_id) (max without (prometheus,receive,instance) ( (time() - cluster_version{type="current"}))) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[7d:12h]) / 60 / 60 / 24',
    },
    {
      name: 'Average subscribed code age (days)',
      query: 'avg_over_time(avg(max by (_id) (max without (prometheus,receive,instance) ( (time() - cluster_version{type="current"}))) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{support=~"Standard|Premium|Layered"})))[7d:12h]) / 60 / 60 / 24',
    },
    {
      name: 'Clusters upgrading to 4.2',
      query: 'count(((count by (_id) (count_over_time(cluster_version{from_version=~"4\\\\.1\\\\.\\\\d+",version=~"4\\\\.2\\\\.\\\\d+",type="updating"}[7d])))*0+1) + on(_id) group_left(_blah) (topk by (_id) (1, 0*subscription_labels{email_domain!~"redhat.com|(.*\\\\.|^)ibm.com"})))',
    },
    {
      name: 'Failed 4.2 upgrades',
      query: 'count(((max by (_id) (sum_over_time((1+0*cluster_version{from_version=~"4\\\\.1\\\\.\\\\d+",version=~"4\\\\.2\\\\.\\\\d+",type="failure"})[7d:15m]))) > 2) + on(_id) group_left(_blah) (topk by (_id) (1, 0*subscription_labels{email_domain!~"redhat.com|(.*\\\\.|^)ibm.com"})))',
    },
    {
      name: 'Clusters upgrading to 4.3',
      query: 'count(((count by (_id) (count_over_time(cluster_version{from_version=~"4\\\\.2\\\\.\\\\d+",version=~"4\\\\.3\\\\.\\\\d+",type="updating"}[7d])))*0+1) + on(_id) group_left(_blah) (topk by (_id) (1, 0*subscription_labels{email_domain!~"redhat.com|(.*\\\\.|^)ibm.com"})))',
    },
    {
      name: 'Failed 4.3 upgrades',
      query: 'count(((max by (_id) (sum_over_time((1+0*cluster_version{from_version=~"4\\\\.2\\\\.\\\\d+",version=~"4\\\\.3\\\\.\\\\d+",type="failure"})[7d:15m]))) > 2) + on(_id) group_left(_blah) (topk by (_id) (1, 0*subscription_labels{email_domain!~"redhat.com|(.*\\\\.|^)ibm.com"})))',
    },
    {
      name: '4.3 clusters',
      query: 'avg_over_time(count(count by (_id) (max without (prometheus,receive,instance) ( cluster_version{type="current",version=~"4\\\\.3\\\\.\\\\d+"} )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[1d:12h])',
    },
    {
      name: '4.2 clusters',
      query: 'avg_over_time(count(count by (_id) (max without (prometheus,receive,instance) ( cluster_version{type="current",version=~"4\\\\.2\\\\.\\\\d+"} )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[1d:12h])',
    },
    {
      name: '4.1 clusters',
      query: 'avg_over_time(count(count by (_id) (max without (prometheus,receive,instance) ( cluster_version{type="current",version=~"4\\\\.1\\\\.\\\\d+"} )) + on (_id) group_left(_blah) (topk by (_id) (1, 0 *subscription_labels{email_domain!~"redhat.com|(^|.*\\\\.)ibm.com"})))[1d:12h])',
    },
  ],
}
