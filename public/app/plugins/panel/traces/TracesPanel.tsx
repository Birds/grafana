import React, { useMemo, useState, createRef } from 'react';
import { PanelProps } from '@grafana/data';
import { TraceView } from 'app/features/explore/TraceView/TraceView';
import { css } from '@emotion/css';
import { transformDataFrames } from 'app/features/explore/TraceView/utils/transform';
import { getDataSourceSrv } from '@grafana/runtime';
import { useAsync } from 'react-use';
import TracePageSearchBar from '@jaegertracing/jaeger-ui-components/src/TracePageHeader/TracePageSearchBar';
import { useSearch } from 'app/features/explore/TraceView/useSearch';
import { TopOfViewRefType } from '@jaegertracing/jaeger-ui-components/src/TraceTimelineViewer/VirtualizedTraceView';

const styles = {
  wrapper: css`
    height: 100%;
    overflow: scroll;
  `,
};

export const TracesPanel: React.FunctionComponent<PanelProps> = ({ data }) => {
  const topOfViewRef = createRef<HTMLDivElement>();
  const traceProp = useMemo(() => transformDataFrames(data.series[0]), [data.series]);
  const { search, setSearch, spanFindMatches } = useSearch(traceProp?.spans);
  const [focusedSpanIdForSearch, setFocusedSpanIdForSearch] = useState('');
  const [searchBarSuffix, setSearchBarSuffix] = useState('');
  const dataSource = useAsync(async () => {
    return await getDataSourceSrv().get(data.request?.targets[0].datasource?.uid);
  });
  const scrollElement = document.getElementsByClassName(styles.wrapper)[0];

  if (!data || !data.series.length || !traceProp) {
    return (
      <div className="panel-empty">
        <p>No data found in response</p>
      </div>
    );
  }

  return (
    <div className={styles.wrapper}>
      <div ref={topOfViewRef}></div>
      <TracePageSearchBar
        navigable={true}
        searchValue={search}
        setSearch={setSearch}
        spanFindMatches={spanFindMatches}
        searchBarSuffix={searchBarSuffix}
        setSearchBarSuffix={setSearchBarSuffix}
        focusedSpanIdForSearch={focusedSpanIdForSearch}
        setFocusedSpanIdForSearch={setFocusedSpanIdForSearch}
      />

      <TraceView
        dataFrames={data.series}
        scrollElement={scrollElement}
        traceProp={traceProp}
        spanFindMatches={spanFindMatches}
        search={search}
        focusedSpanIdForSearch={focusedSpanIdForSearch}
        queryResponse={data}
        datasource={dataSource.value}
        topOfViewRef={topOfViewRef}
        topOfViewRefType={TopOfViewRefType.Panel}
      />
    </div>
  );
};
