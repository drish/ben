'use strict'

const Table = require('cli-table2')
const { Suite } = require('benchmark')
const sortBy = require('sort-array')
const arraySort = require('array-sort')

const curr = require('sort-arr')

const bench = new Suite()

const bar = 'band'
const foo = [
  { band: 'Lights', members: 1 },
  { band: 'Blink-182', members: 3 },
  { band: 'Jamestown Story', members: 1 },
  { band: 'Linkin Park', members: 1 },
  { band: 'Aerosmith', members: 3 },
  { band: 'Guns n Roses', members: 1 },
  { band: 'Priest', members: 1 },
  { band: 'PVRIS', members: 3 },
  { band: 'Yellowcard', members: 1 },
  { band: 'City Lights', members: 1 },
  { band: 'Anberlin', members: 3 },
  { band: 'The Red Jumpsuit Apparatus', members: 1 },
  { band: 'Airspoken', members: 1 },
  { band: 'Amycambe', members: 3 },
  { band: 'Plus 44', members: 1 },
  { band: 'Box Car Racer', members: 1 },
  { band: 'Sum 41', members: 3 },
  { band: 'Oasis', members: 1 }
]

bench
  .add('sort-array', () => sortBy(foo, bar))
  .add('array-sort', () => arraySort(foo, bar))
  .add('sort-arr', () => curr(foo, bar))
  .on('cycle', e => console.log(String(e.target)))
  .on('complete', function() {
    console.log('Fastest is ' + this.filter('fastest').map('name'))

    const tbl = new Table({
      head: ['Name', 'Mean time', 'Ops/sec', 'Diff']
    })

    let prev
    let diff

    bench.forEach(el => {
      if (prev) {
        diff = ((el.hz - prev) * 100 / prev).toFixed(2) + '% faster'
      } else {
        diff = 'N/A'
      }
      prev = el.hz
      tbl.push([el.name, el.stats.mean, el.hz.toLocaleString(), diff])
    })
    console.log(tbl.toString())
  })
  .run()
